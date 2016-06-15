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
	name := m.ServiceName() + "-v" + m.BuildVersion()
	services, _, err := consul.Health().Service(name, "", true, nil)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return fmt.Errorf("Service `%s` not started!", name)
	} else {
		log.Printf("Service `%s` started with %v instances.", name, len(services))
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
		Key:   fmt.Sprintf("services/routes/%s", name),
		Value: routesJson,
	}, nil); err != nil {
		return err
	}

	log.Println(color.GreenString("Service `%s` released with routes: %s", name, string(routesJson)))

	// find old services with this routes
	kvPairs, _, err := consul.KV().List(fmt.Sprintf("services/routes/%s-v", m.ServiceName()), nil)
	if err != nil {
		return err
	}

	for _, kv := range kvPairs {
		if !strings.Contains(kv.Key, name) { // skip current service
			oldRoutes := make([]map[string]string, 0)
			if err := json.Unmarshal(kv.Value, &oldRoutes); err != nil {
				return err
			}

			for _, route := range routes {
				for _, oldRoute := range oldRoutes {
					if utils.MapsEqual(route, oldRoute) {
						oldName := strings.TrimPrefix(kv.Key, "services/routes/")
						log.Printf("Found %s with routes %v. Remove it!", oldName, oldRoute)

						if _, err := consul.KV().Delete(kv.Key, nil); err != nil {
							return err
						}

						log.Printf("Delete %s from marathon after 3 minutes...", oldName)

						<-time.NewTimer(time.Minute * 3).C
						log.Printf("Delete %s from marathon", oldName)

						marathonApi := serveMarathon.MarathonClient(m)
						if _, err := marathonApi.DeleteApplication(m.GetStringOr("info.category", "") + "/" + oldName); err != nil {
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
