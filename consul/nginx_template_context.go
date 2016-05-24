package consul

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"
)

var upstreamNameRegex = regexp.MustCompile("[^\\w]+")

func NginxTemplateContextCommand() cli.Command {
	return cli.Command{
		Name:  "nginx-template-context",
		Usage: "Collect and return data for consul-template",
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			upstreams := make(map[string][]map[string]interface{})
			servers := make(map[string]map[string]map[string]string)

			allRoutes, _, err := consul.KV().List("services/routes/", nil)
			if err != nil {
				return err
			}

			for _, kv := range allRoutes {
				name := strings.TrimPrefix(kv.Key, "services/routes/")
				upstream := upstreamNameRegex.ReplaceAllString("serve_"+name, "_")

				routes := make([]map[string]string, 0)
				if err := json.Unmarshal(kv.Value, &routes); err != nil {
					return err
				}

				instances, _, err := consul.Health().Service(name, "", true, nil)
				if err != nil {
					panic(err)
				}

				if len(instances) == 0 {
					break
				}

				if _, ok := upstreams[upstream]; !ok {
					upstreams[upstream] = make([]map[string]interface{}, 0)
				}

				for _, inst := range instances {
					address := inst.Node.Address
					if inst.Service.Address != "" {
						address = inst.Service.Address
					}

					upstreams[upstream] = append(upstreams[upstream], map[string]interface{}{
						"address": address,
						"port":    inst.Service.Port,
					})
				}

				for _, route := range routes {
					location, ok := route["location"]
					if !ok {
						location = "/"
					}

					staging, ok := route["staging"]
					if !ok {
						staging = "live"
					}

					if _, ok := servers[route["host"]]; !ok {
						servers[route["host"]] = make(map[string]map[string]string, 0)
					}

					if _, ok := servers[route["host"]][location]; !ok {
						servers[route["host"]][location] = make(map[string]string, 0)
					}

					if _, ok := servers[route["host"]][location][staging]; !ok {
						servers[route["host"]][location][staging] = upstream
					}
				}
			}

			out, _ := json.MarshalIndent(map[string]interface{}{
				"upstreams": upstreams,
				"servers":   servers,
			}, "", "  ")

			fmt.Fprintln(os.Stdout, string(out))
			return nil
		},
	}
}
