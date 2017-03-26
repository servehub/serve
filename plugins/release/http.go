package release

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/fatih/color"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("release.http", ReleaseHttp{})
}

type ReleaseHttp struct{}

func (p ReleaseHttp) Run(data manifest.Manifest) error {
	if !data.Has("routes") {
		log.Println("No routes configured for release.")
		return nil
	}

	consul, err := utils.ConsulClient(data.GetString("consul-address"))
	if err != nil {
		return err
	}

	fullName := data.GetString("full-name")

	// check current service is alive and healthy
	if err := backoff.Retry(func() error {
		services, _, err := consul.Health().Service(fullName, "", true, nil)
		if err != nil {
			log.Println(color.RedString("Error in check health in consul: %v", err))
			return err
		}

		if len(services) == 0 {
			log.Printf("Service `%s` not started yet! Retry...", fullName)
			return fmt.Errorf("Service `%s` not started!", fullName)
		} else {
			log.Printf("Service `%s` started with %v instances.", fullName, len(services))
			return nil
		}
	}, backoff.NewExponentialBackOff()); err != nil {
		return err
	}

	routeVars := make(map[string]string, 0)
	if data.Has("route-vars") {
		if err := json.Unmarshal([]byte(data.GetString("route-vars")), &routeVars); err != nil {
			log.Println(color.RedString("Error parse route json: %v, %s", err, data.GetString("route")))
			return err
		}
	}

	if data.Has("stage") {
		routeVars["stage"] = data.GetString("stage")
	}

	// collect routes
	routes := consulRoutes{}
	for _, route := range data.GetArray("routes") {
		if !route.Has("host") {
			log.Printf("Not found 'host': %s, skip...", route.String())
			continue
		}

		fields := make(map[string]string)
		for k, v := range route.Unwrap().(map[string]interface{}) {
			if k != "host" && k != "location" {
				fields[k] = fmt.Sprintf("%v", v)
			}
		}

		routes.Routes = append(routes.Routes, consulRoute{
			Host:     route.GetString("host"),
			Location: route.GetStringOr("location", ""),
			Vars:     utils.MergeMaps(fields, routeVars),
		})
	}

	if len(routes.Routes) == 0 {
		log.Println("No routes configured for release.")
		return nil
	}

	routesJson, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return err
	}

	// write routes to consul kv
	if err := utils.PutConsulKv(consul, "services/routes/"+fullName, string(routesJson)); err != nil {
		return err
	}

	utils.DelConsulKv(consul, "services/outdated/"+fullName) // force delete outdated key (if exists)

	log.Println(color.GreenString("Service `%s` released with routes: %s", fullName, string(routesJson)))

	// find old services with the same routes
	existsRoutes, err := utils.ListConsulKv(consul, "services/routes/"+data.GetString("name-prefix"), nil)
	if err != nil {
		return err
	}

	for _, existsRoute := range existsRoutes {
		if existsRoute.Key != fmt.Sprintf("services/routes/%s", fullName) { // skip current service
			oldRoutes := consulRoutes{}
			if err := json.Unmarshal(existsRoute.Value, &oldRoutes); err != nil {
				return err
			}

		OuterLoop:
			for _, route := range routes.Routes {
				for _, oldRoute := range oldRoutes.Routes {
					if route.Host == oldRoute.Host && route.Location == oldRoute.Location && utils.MapsEqual(route.Vars, oldRoute.Vars) {
						outdated := strings.TrimPrefix(existsRoute.Key, "services/routes/")
						log.Println(color.GreenString("Found %s with the same routes %v. Remove it!", outdated, string(existsRoute.Value)))

						if err := utils.DelConsulKv(consul, existsRoute.Key); err != nil {
							return err
						}

						if err := utils.MarkAsOutdated(consul, outdated, time.Duration(data.GetIntOr("outdated-timeout-sec", 600))*time.Second); err != nil {
							return err
						}

						break OuterLoop
					}
				}
			}
		}
	}

	return nil
}

type consulRoutes struct {
	Routes []consulRoute `json:"routes"`
}

type consulRoute struct {
	Host     string            `json:"host"`
	Location string            `json:"location,omitempty"`
	Vars     map[string]string `json:"vars,omitempty"`
}
