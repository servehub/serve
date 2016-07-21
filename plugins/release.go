package plugins

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/fatih/color"
	consul "github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("release", Release{})
}

type Release struct{}

func (p Release) Run(data manifest.Manifest, vars map[string]string) error {
	log.Println(color.BlueString("data = %s", data))
	consulApi, err := ConsulClient(data.GetString("consul_host"))
	if err != nil {
		return err
	}

	fullName := data.GetString("full-name")

	// check current service is alive and healthy
	if err := backoff.Retry(func() error {
		services, _, err := consulApi.Health().Service(fullName, "", true, nil)
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
			return fmt.Errorf("'host' is required for routes! Given: %s", route.String())
		}

		fields := make(map[string]string)
		for k, v := range route.Unwrap().(map[string]interface{}) {
			fields[k] = fmt.Sprintf("%v", v)
		}

		routes = append(routes, utils.MergeMaps(fields, routeVars))
	}

	routesJson, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return err
	}

	// write routes to consul kv
	if _, err := consulApi.KV().Put(&consul.KVPair{
		Key:   fmt.Sprintf("services/routes/%s", fullName),
		Value: routesJson,
	}, nil); err != nil {
		return err
	}

	log.Println(color.GreenString("Service `%s` released with routes: %s", fullName, string(routesJson)))

	// find old services with the same routes
	existsRoutes, _, err := consulApi.KV().List(fmt.Sprintf("services/routes/%s", data.GetString("name-prefix")), nil)
	if err != nil {
		return err
	}

	for _, existsRoute := range existsRoutes {
		if existsRoute.Key != fmt.Sprintf("services/routes/%s", fullName) { // skip current service
			oldRoutes := make([]map[string]string, 0)
			if err := json.Unmarshal(existsRoute.Value, &oldRoutes); err != nil {
				return err
			}

			OuterLoop: for _, route := range routes {
				for _, oldRoute := range oldRoutes {
					if utils.MapsEqual(route, oldRoute) {
						outdated := strings.TrimPrefix(existsRoute.Key, "services/routes/")
						log.Println(color.GreenString("Found %s with the same routes %v. Remove it!", outdated, string(existsRoute.Value)))

						if _, err := consulApi.KV().Delete(existsRoute.Key, nil); err != nil {
							return err
						}

						if data.Has("outdated.marathon") {
							delay := data.GetInt("outdated.marathon.delay-minutes")
							log.Printf("Delete %s from marathon after %s minutes...", outdated, delay)

							marathonApi, err := MarathonClient(data.GetString("outdated.marathon.marathon-host"))
							if err != nil {
								return err
							}

							<-time.NewTimer(time.Duration(delay) * time.Minute).C
							log.Printf("Delete %s from marathon", outdated)

							if _, err := marathonApi.DeleteApplication(outdated, true); err != nil {
								log.Println(color.RedString("Error on delete old instance: %v", err))
								return err
							}
						}

						break OuterLoop
					}
				}
			}
		}
	}

	return nil
}
