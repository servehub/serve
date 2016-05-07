package consul

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"
)

var upstreamNameRegex = regexp.MustCompile("[^\\w]+")

func NginxTemplateContextCommand() cli.Command {
	return cli.Command{
		Name:  "nginx-tempalte-context",
		Usage: "Collect and return data for consul-tempalte",
		Action: func(c *cli.Context) {
			consul, _ := api.NewClient(api.DefaultConfig())

			upstreams := make(map[string][]map[string]interface{})
			servers := make(map[string]map[string]map[string]string)

			allServices, _, err := consul.Catalog().Services(&api.QueryOptions{})
			if err != nil {
				panic(err)
			}

			for s, allTags := range allServices {
				if _, ok := ParseTags(allTags)["domain"]; ok {
					services, _, err := consul.Health().Service(s, "", true, &api.QueryOptions{})
					if err != nil {
						panic(err)
					}

					for _, serv := range services {
						tags := ParseTags(serv.Service.Tags)

						address := serv.Node.Address
						if serv.Service.Address != "" {
							address = serv.Service.Address
						}

						location, ok := tags["location"]
						if !ok {
							location = "/"
						}

						staging, ok := tags["staging"]
						if !ok {
							staging = "live"
						}

						upstream := upstreamNameRegex.ReplaceAllString("ups_"+tags["domain"]+"_"+location+"_"+staging, "_")

						if _, ok := upstreams[upstream]; !ok {
							upstreams[upstream] = make([]map[string]interface{}, 0)
						}

						upstreams[upstream] = append(upstreams[upstream], map[string]interface{}{
							"address": address,
							"port":    serv.Service.Port,
						})

						if _, ok := servers[tags["domain"]]; !ok {
							servers[tags["domain"]] = make(map[string]map[string]string, 0)
						}

						if _, ok := servers[tags["domain"]][location]; !ok {
							servers[tags["domain"]][location] = make(map[string]string, 0)
						}

						if _, ok := servers[tags["domain"]][location][staging]; !ok {
							servers[tags["domain"]][location][staging] = upstream
						}
					}
				}
			}

			out, _ := json.MarshalIndent(map[string]interface{}{
				"upstreams": upstreams,
				"servers":   servers,
			}, "", "  ")

			fmt.Fprintln(os.Stdout, string(out))
			os.Exit(0)
		},
	}
}
