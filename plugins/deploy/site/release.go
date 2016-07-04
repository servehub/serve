package site

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"

	serveConsul "github.com/InnovaCo/serve/consul"
	"github.com/InnovaCo/serve/manifest"
	serveMarathon "github.com/InnovaCo/serve/marathon"
	"github.com/InnovaCo/serve/utils"
)

type SiteRelease struct{}

func (_ SiteRelease) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	consul := serveConsul.ConsulClient(m)

	// check current service is alive
	fullName := m.ServiceFullNameWithVersion("/")
	services, _, err := consul.Health().Service(fullName, "", true, nil)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return fmt.Errorf("Service `%s` not started!", fullName)
	} else {
		log.Printf("Service `%s` started with %v instances.", fullName, len(services))
	}

	routeFlags := make(map[string]string, 0)
	if m.Args("route") != "" {
		if err := json.Unmarshal([]byte(m.Args("route")), &routeFlags); err != nil {
			return err
		}
	}

	// collect routes
	routes := make([]map[string]string, 0)
	for _, route := range sub.Array("routes") {
		if route.GetBool("featured") == (m.Args("feature") != "") {
			routes = append(routes, utils.MergeMaps(map[string]string{
				"host":     route.Template("host"),
				"location": route.TemplateOr("location", "/"),
			}, routeFlags))
		}
	}

	routesJson, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return err
	}

	// write routes to consul kv
	if _, err := consul.KV().Put(&api.KVPair{
		Key:   fmt.Sprintf("services/routes/%s", fullName),
		Value: routesJson,
	}, nil); err != nil {
		return err
	}

	log.Println(color.GreenString("Service `%s` released with routes: %s", fullName, string(routesJson)))

	// find old services with this routes
	routesData, _, err := consul.KV().List(fmt.Sprintf("services/routes/%s-v", m.ServiceFullName("/")), nil)
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

						if _, err := consul.KV().Delete(existRoute.Key, nil); err != nil {
							return err
						}

						log.Printf("Delete %s from marathon after 3 minutes...", oldName)

						<-time.NewTimer(time.Minute * 3).C
						log.Printf("Delete %s from marathon", oldName)

						marathonApi := serveMarathon.MarathonClient(m)
						if _, err := marathonApi.DeleteApplication(oldName); err != nil {
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
