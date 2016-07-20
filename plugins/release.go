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

	// check current service is alive
	fullName := nameEscapeRegex.ReplaceAllString(data.GetString("full_name_version"), "-")
	services, _, err := consulApi.Health().Service(fullName, "", true, nil)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return fmt.Errorf("Service `%s` not started!", fullName)
	} else {
		log.Printf("Service `%s` started with %v instances.", fullName, len(services))
	}

	routeFlags := make(map[string]string, 0)
	if data.GetString("route") != "" {
		if err := json.Unmarshal([]byte(data.GetString("route")), &routeFlags); err != nil {
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

	// write routes to consul kv
	if _, err := consulApi.KV().Put(&consul.KVPair{
		Key:   fmt.Sprintf("services/routes/%s", fullName),
		Value: routesJson,
	}, nil); err != nil {
		return err
	}

	log.Println(color.GreenString("Service `%s` released with routes: %s", fullName, string(routesJson)))

	// find old services with this routes
	routesData, _, err := consulApi.KV().List(fmt.Sprintf("services/routes/%s-v", nameEscapeRegex.ReplaceAllString(data.GetString("full_name"), "-")), nil)
	if err != nil {
		return err
	}

	for _, existRoute := range routesData {
		if !strings.Contains(existRoute.Key, fullName) { // skip current service
			oldRoutes := make([]map[string]string, 0)
			if err := json.Unmarshal(existRoute.Value, &oldRoutes); err != nil {
				return err
			}

			for _, route := range routes {
				for _, oldRoute := range oldRoutes {
					if utils.MapsEqual(route, oldRoute) {
						oldName := strings.TrimPrefix(existRoute.Key, "services/routes/")
						log.Printf("Found %s with routes %v. Remove it!", oldName, oldRoute)

						if _, err := consulApi.KV().Delete(existRoute.Key, nil); err != nil {
							return err
						}

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

						return nil
					}
				}
			}
		}
	}

	return nil
}
