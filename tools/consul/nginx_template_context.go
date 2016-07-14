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
			services := make(map[string]map[string]map[string]map[string]string)

			allRoutes, _, err := consul.KV().List("services/routes/", nil)
			if err != nil {
				return fmt.Errorf("Error on load routes: %s", err)
			}

			for _, kv := range allRoutes {
				name := strings.TrimPrefix(kv.Key, "services/routes/")
				upstream := ""

				routes := make([]map[string]string, 0)
				if err := json.Unmarshal(kv.Value, &routes); err != nil {
					return fmt.Errorf("Error on parse route json: %s. Serive `%s`, json: %s", err, name, string(kv.Value))
				}

				instances, _, err := consul.Health().Service(name, "", true, nil)
				if err != nil {
					return fmt.Errorf("Error on get service `%s` health: %s", name, err)
				}

				if len(instances) == 0 {
					break
				}

				for _, inst := range instances {
					if inst.Service.Port != 0 {
						upstream = upstreamNameRegex.ReplaceAllString("serve_"+name, "_")

						if _, ok := upstreams[upstream]; !ok {
							upstreams[upstream] = make([]map[string]interface{}, 0)
						}

						address := inst.Node.Address
						if inst.Service.Address != "" {
							address = inst.Service.Address
						}

						upstreams[upstream] = append(upstreams[upstream], map[string]interface{}{
							"address": address,
							"port":    inst.Service.Port,
						})
					}
				}

				for _, route := range routes {
					location, ok := route["location"]
					if !ok {
						location = "/"
					}

					stage, ok := route["stage"]
					if !ok {
						stage = "live"
					}

					if _, ok := services[route["host"]]; !ok {
						services[route["host"]] = make(map[string]map[string]map[string]string, 0)
					}

					if _, ok := services[route["host"]][location]; !ok {
						services[route["host"]][location] = make(map[string]map[string]string, 0)
					}

					if _, ok := services[route["host"]][location][stage]; !ok {
						services[route["host"]][location][stage] = map[string]string{
							"upstream": upstream,
							"service": name,
						}
					}
				}
			}

			out, _ := json.MarshalIndent(map[string]interface{}{
				"upstreams": upstreams,
				"services":  services,
			}, "", "  ")

			fmt.Fprintln(os.Stdout, string(out))
			return nil
		},
	}
}
