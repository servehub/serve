package consul

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/utils"
)

func RouteCommand() cli.Command {
	return cli.Command{
		Name:  "route",
		Usage: "Save app route to consul",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "service"},
			cli.StringFlag{Name: "host"},
			cli.StringFlag{Name: "location", Value: "/"},
			cli.StringFlag{Name: "route"},
		},
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			name := c.String("service")
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
			if c.IsSet("route") {
				if err := json.Unmarshal([]byte(c.String("route")), &routeFlags); err != nil {
					return err
				}
			}

			route := utils.MergeMaps(map[string]string{
				"host":     c.String("host"),
				"location": c.String("location"),
			}, routeFlags)

			routesJson, err := json.MarshalIndent([]map[string]string{route}, "", "  ")
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

			log.Println(color.GreenString("Updated routes for `%s`: %s", name, string(routesJson)))

			return nil
		},
	}
}
