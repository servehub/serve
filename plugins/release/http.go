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

var routeVarsExclude = []string{"host", "location", "cache", "extra", "ssl", "stripPrefix", "redirectHttps", "enabled", "hostAliases"}

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

	defaults := data.GetTree("defaults").Unwrap().(map[string]interface{})

	// collect routes
	routes := consulRoutes{}
	for _, route := range data.GetArray("routes") {

		// set default values
		for k, v := range defaults {
			if !route.Has(k) {
				route.Set(k, v)
			}
		}

		if !route.Has("host") || !route.GetBool("enabled") {
			log.Printf("Not found 'host' or host disabled: %s, skip...", route.String())
			continue
		}

		fields := make(map[string]string)
		for k, v := range route.Unwrap().(map[string]interface{}) {
			if !utils.Contains(k, routeVarsExclude) {
				fields[k] = fmt.Sprintf("%v", v)
			}
		}

		var cache map[string]interface{} = nil
		if route.Has("cache") {
			cache = route.GetTree("cache").Unwrap().(map[string]interface{})
		}

		var params = make(map[string]interface{})

		if route.GetBool("stripPrefix") {
			params["stripPrefix"] = true
		}

		if route.GetBool("redirectHttps") {
			params["redirectHttps"] = true
		}

		if route.Has("hostAliases") {
			aliases := ""
			for _, v := range data.GetArrayForce("hostAliases") {
				aliases = fmt.Sprintf("%s %s", aliases, v)
			}
			params["hostAliases"] = aliases
		}

		var ssl map[string]interface{} = nil
		if route.Has("ssl") {
			ssl = route.GetTree("ssl").Unwrap().(map[string]interface{})
		}

		extra := route.GetStringOr("extra", "")
		if data.GetBool("maintenance") {
			extra += "\n return 503; \n"
		}

		routes.Routes = append(routes.Routes, consulRoute{
			Host:     route.GetString("host"),
			Location: route.GetStringOr("location", ""),
			Vars:     utils.MergeMaps(fields, routeVars),
			Cache:    cache,
			Ssl:      ssl,
			Extra:    extra,
			Params:   params,
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
					delete(route.Vars, "public") // ignore public filter
					delete(oldRoute.Vars, "public")

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
	Host     string                 `json:"host"`
	Location string                 `json:"location,omitempty"`
	Vars     map[string]string      `json:"vars,omitempty"`
	Cache    map[string]interface{} `json:"cache,omitempty"`
	Ssl      map[string]interface{} `json:"ssl,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Extra    string                 `json:"extra,omitempty"`
}
