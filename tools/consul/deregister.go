package consul

import (
	"log"
	"strings"
	"encoding/json"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"
)

func DeregisterCommand() cli.Command {
	return cli.Command{
		Name:  "deregister",
		Usage: "Deregister all unhealth services on cluster",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "dry-run"},
		},
		Subcommands: []cli.Command{
			{
				Name:  "service",
				Action: func(c *cli.Context) error {
					consul, _ := api.NewClient(api.DefaultConfig())

					name := c.Args().First()
					if name == "" {
						return f
					}

					servs, err := consul.Agent().Services()
					if err != nil {
						return err
					}

					js, _ := json.MarshalIndent(servs, "", "  ")
					println(string(js))

					return nil
				},
			},
			{
				Name:  "unhealth",
				Usage: "Deregister all unhealth services on cluster",
				Action: func(c *cli.Context) error {
					consul, _ := api.NewClient(api.DefaultConfig())

					checks, err := consul.Agent().Checks()
					if err != nil {
						return err
					}

					js, _ := json.MarshalIndent(checks, "", "  ")
					println(string(js))

					name := c.String("service")

					for _, check := range checks {
						if check.Status == "critical" || (name != "" && strings.Contains(check.ServiceID, name)) {
							log.Println("Deregistering", check.ServiceID, "=", check.Status)

							if !c.GlobalBool("dry-run") {
								err := consul.Agent().ServiceDeregister(check.ServiceID)
								if err != nil {
									return err
								}
							}
						}
					}

					return nil
				},
			},
		},
	}
}
