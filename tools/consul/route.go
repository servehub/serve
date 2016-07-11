package consul

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"
	"github.com/cenk/backoff"
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

			err := backoff.Retry(func() error {
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
			}, backoff.NewExponentialBackOff())

			if err != nil {
				return err
			}

			routes := c.String("routes")

			if err != nil {
				return err
			}

			// write routes to consul kv
			if _, err := consul.KV().Put(&api.KVPair{
				Key:   fmt.Sprintf("services/routes/%s", name),
				Value: []byte(routes),
			}, nil); err != nil {
				return err
			}

			log.Println(color.GreenString("Updated routes for `%s`: %s", name, string(routes)))

			return nil
		},
	}
}
