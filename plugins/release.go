package plugins

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/fatih/color"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("release", Release{})
}

type Release struct{}

func (p Release) Run(data manifest.Manifest) error {
	if !data.Has("routes") {
		log.Println("No routes configured for release.")
		return nil
	}

	consul, err := ConsulClient(data.GetString("consul-host"))
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
	if data.Has("route") {
		if err := json.Unmarshal([]byte(data.GetString("route")), &routeVars); err != nil {
			log.Println(color.RedString("Error parse route json: %v, %s", err, data.GetString("route")))
			return err
		}
	}

	// collect routes
	routes := make([]map[string]string, 0)
	for _, route := range data.GetArray("routes") {
		if !route.Has("host") {
			log.Printf("Not found 'host': %s, skip...", route.String())
			continue
		}

		fields := make(map[string]string)
		for k, v := range route.Unwrap().(map[string]interface{}) {
			fields[k] = fmt.Sprintf("%v", v)
		}

		routes = append(routes, utils.MergeMaps(fields, routeVars))
	}

	if len(routes) == 0 {
		log.Println("No routes configured for release.")
		return nil
	}

	routesJson, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return err
	}

	// write routes to consul kv
	if err := putConsulKv(consul, "services/routes/"+fullName, string(routesJson)); err != nil {
		return err
	}

	log.Println(color.GreenString("Service `%s` released with routes: %s", fullName, string(routesJson)))

	// find old services with the same routes
	existsRoutes, err := listConsulKv(consul, "services/routes/"+data.GetString("name-prefix"), nil)
	if err != nil {
		return err
	}

	for _, existsRoute := range existsRoutes {
		if existsRoute.Key != fmt.Sprintf("services/routes/%s", fullName) { // skip current service
			oldRoutes := make([]map[string]string, 0)
			if err := json.Unmarshal(existsRoute.Value, &oldRoutes); err != nil {
				return err
			}

		OuterLoop:
			for _, route := range routes {
				for _, oldRoute := range oldRoutes {
					if utils.MapsEqual(route, oldRoute) {
						outdated := strings.TrimPrefix(existsRoute.Key, "services/routes/")
						log.Println(color.GreenString("Found %s with the same routes %v. Remove it!", outdated, string(existsRoute.Value)))

						if err := delConsulKv(consul, existsRoute.Key); err != nil {
							return err
						}

						outdatedJson := fmt.Sprintf(`{"endOfLife":%d}`, time.Now().UnixNano()/int64(time.Millisecond))
						if err := putConsulKv(consul, "services/outdated/"+outdated, outdatedJson); err != nil {
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
