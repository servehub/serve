package plugins

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	consul "github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
	"github.com/cenk/backoff"
)

func init() {
	manifest.PluginRegestry.Add("release", Release{})
}

type Release struct{}

func (p Release) Run(data manifest.Manifest) error {
	consulApi, err := ConsulClient(data.GetString("consul_host"))
	if err != nil {
		return err
	}

	// check current service is alive
	fullName := nameEscapeRegex.ReplaceAllString(data.GetString("full_name_version"), "-")

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
		return  err
	}

	routeFlags := make(map[string]string, 0)
	if data.GetString("route") != "" {
		if err := json.Unmarshal([]byte(data.GetString("route")), &routeFlags); err != nil {
			log.Println(color.RedString("Error parse routes json: %v, %s", err))
			return err
		}
	}

	// collect routes
	routes := make([]map[string]string, 0)
	for _, route := range data.GetArray("routes") {
		routes = append(routes, utils.MergeMaps(map[string]string{
			"host":     route.GetString("host"),
			"location": route.GetStringOr("location", "/"),
		}, routeFlags))
	}

	routesJson, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return err
	}

	// find old services with this routes
	routesData, _, err := consulApi.KV().List("services/routes/", nil)
	if err != nil {
		return err
	}

	// write routes to consul kv
	if _, err := consulApi.KV().Put(&consul.KVPair{
		Key:   fmt.Sprintf("services/routes/%s", fullName),
		Value: routesJson,
	}, nil); err != nil {
		log.Println(color.RedString("Error save routes to consul: %v", err))
		return err
	}

	log.Println(color.GreenString("Service `%s` released with routes: %s", fullName, string(routesJson)))

	for _, existRoute := range routesData {
		if existRoute.Key != fmt.Sprintf("services/routes/%s", fullName) {
			oldRoutes := make([]map[string]string, 0)
			if err := json.Unmarshal(existRoute.Value, &oldRoutes); err != nil {
				return err
			}
			OuterLoop: for _, route := range routes {
				for _, oldRoute := range oldRoutes {
					if utils.MapsEqual(route, oldRoute) {
						oldName := strings.TrimPrefix(existRoute.Key, "services/routes/")
						log.Printf("Found %s with routes %v. Remove it!", oldName, oldRoute)

						if _, err := consulApi.KV().Delete(existRoute.Key, nil); err != nil {
							return err
						}

						if (data.Has("marathon")) {
							log.Printf("Delete %s from marathon after 3 minutes...", oldName)

							<-time.NewTimer(time.Minute * 3).C
							log.Printf("Delete %s from marathon", oldName)

							marathonApi, err := MarathonClient(data.GetString("marathon_host"))
							if err != nil {
								return err
							}

							if _, err := marathonApi.DeleteApplication(oldName, true); err != nil {
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
