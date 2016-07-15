package consul

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cenk/backoff"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"
	"github.com/InnovaCo/serve/utils"
	"strings"
)

func RouteCommand() cli.Command {
	return cli.Command{
		Name:  "route",
		Usage: "Save app route to consul",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "service"},
			cli.StringFlag{Name: "routes"},
		},
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			name := c.String("service")

			if err := backoff.Retry(func() error {
				services, _, err := consul.Health().Service(name, "", true, nil)
				if err != nil {
					log.Println(color.RedString("Error in check health in consul: %v", err))
					return err
				}

				if len(services) == 0 {
					log.Printf("Service `%s` not started yet! Retry...", name)
					return fmt.Errorf("Service `%s` not started!", name)
				} else {
					log.Printf("Service `%s` started with %v instances.", name, len(services))
					return nil
				}
			}, backoff.NewExponentialBackOff()); err != nil {
				return err
			}

			routes := make([]map[string]string, 0)
			if err := json.Unmarshal([]byte(c.String("routes")), &routes); err != nil {
				log.Println(color.RedString("Error parse routes json: %v, %s", err, c.String("routes")))
				return err
			}

			routesJson, _ := json.MarshalIndent(routes, "", "  ")

			// find old services with this routes
			routesData, _, err := consul.KV().List("services/routes/", nil)
			if err != nil {
				return  err
			}

			// write routes to consul kv
			if _, err := consul.KV().Put(&api.KVPair{
				Key:   fmt.Sprintf("services/routes/%s", name),
				Value: routesJson,
			}, nil); err != nil {
				log.Println(color.RedString("Error save routes to consul: %v", err))
				return err
			}

			log.Println(color.RedString("routesData = `%s`", routesData))
			for _, existRoute := range routesData {
				if !strings.Contains(existRoute.Key, name) {
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
							}
						}
					}
				}
			}

			log.Println(color.GreenString("Updated routes for `%s`: %s", name, string(routesJson)))

			return nil
		},
	}
}
