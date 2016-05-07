package marathon

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"

	consulCmd "github.com/kulikov/serve/consul"
	"github.com/kulikov/serve/utils"
)

func ReleaseSiteCommand() cli.Command {
	return cli.Command{
		Name: "release-site",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "marathon"},
			cli.StringFlag{Name: "name"},
			cli.StringFlag{Name: "version"},
		},
		Action: func(c *cli.Context) {
			consul, _ := api.NewClient(api.DefaultConfig())

			if staged, _, err := consul.Catalog().Service(c.GlobalString("name"), "version:"+c.GlobalString("version"), &api.QueryOptions{}); err == nil {
				for _, serv := range staged {
					if _, err := consul.Catalog().Register(&api.CatalogRegistration{
						Node:    serv.Node,
						Address: serv.Address,
						Service: &api.AgentService{
							ID:                serv.ServiceID,
							Service:           serv.ServiceName,
							Tags:              consulCmd.MapToList(utils.MergeMaps(consulCmd.ParseTags(serv.ServiceTags), consulCmd.TagsFromFlags(c))),
							Port:              serv.ServicePort,
							Address:           serv.ServiceAddress,
							EnableTagOverride: serv.ServiceEnableTagOverride,
						},
						Check: &api.AgentCheck{
							Node:        serv.Node,
							CheckID:     "service:" + serv.ServiceID,
							ServiceID:   serv.ServiceID,
							ServiceName: serv.ServiceName,
							Status:      api.HealthPassing,
						},
					}, &api.WriteOptions{}); err != nil {
						log.Println("Register:", err)
					}
				}
			}
		},
	}
}
