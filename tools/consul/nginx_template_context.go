package consul

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/utils"
)

var upstreamNameRegex = regexp.MustCompile("[^\\w]+")

func NginxTemplateContextCommand() cli.Command {
	return cli.Command{
		Name:  "nginx-template-context",
		Usage: "Collect and return data for consul-template",
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			upstreams := make(map[string][]map[string]interface{})
			services := make(map[string]map[string][]map[string]string)
			localServers := make(map[string]map[string]string)
			duplicates := make(map[string]string)

			allRoutes, _, err := consul.KV().List("services/routes/", nil)
			if err != nil {
				return fmt.Errorf("Error on load routes: %s", err)
			}

			for _, kv := range allRoutes {
				name := strings.TrimPrefix(kv.Key, "services/routes/")
				upstream := ""

				instances, _, err := consul.Health().Service(name, "", true, nil)
				if err != nil {
					return fmt.Errorf("Error on get service `%s` health: %s", name, err)
				}

				if len(instances) == 0 {
					break
				}

				routes := make([]map[string]string, 0)
				if err := json.Unmarshal(kv.Value, &routes); err != nil {
					return fmt.Errorf("Error on parse route json: %s. Serive `%s`, json: %s", err, name, string(kv.Value))
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
					host, ok := route["host"]
					if !ok {
						return fmt.Errorf("Host is required for routing! Service `%s`", name)
					}

					location, ok := route["location"]
					if !ok {
						location = "/"
					}

					routeUpstream := upstream
					if ups, ok := route["upstream"]; ok && ups == "local" { // so far support only "local" custom upstream
						routeUpstream = ""
					}

					if routeUpstream == "" {
						localServers[name] = map[string]string{
							"package": name,
						}
					}

					delete(route, "host")
					delete(route, "location")
					delete(route, "upstream")

					if _, ok := services[host]; !ok {
						services[host] = make(map[string][]map[string]string, 0)
					}

					if _, ok := services[host][location]; !ok {
						services[host][location] = make([]map[string]string, 0)
					}

					routeKeys := "-"
					routeValues := "-"
					for k, v := range route {
						routeKeys += "${" + k + "}-"
						routeValues += v + "-"
					}

					if exists, ok := duplicates[host+location+routeKeys+routeValues]; !ok {
						duplicates[host+location+routeKeys+routeValues] = name
					} else {
						fmt.Fprintln(os.Stderr, color.RedString("Service with the same routes already exists! exists: %s, skipped: %s", exists, name))
						continue
					}

					services[host][location] = append(services[host][location], map[string]string{
						"upstream":    routeUpstream,
						"package":     name,
						"routeKeys":   routeKeys,
						"routeValues": routeValues,
						"sortIndex":   strconv.Itoa(len(route)),
					})
				}
			}

			// sort routes by sort index
			for _, hh := range services {
				for _, ll := range hh {
					sort.Sort(utils.BySortIndex(ll))
				}
			}

			out, _ := json.MarshalIndent(map[string]interface{}{
				"upstreams":    upstreams,
				"services":     services,
				"localServers": localServers,
			}, "", "  ")

			fmt.Fprintln(os.Stdout, string(out))
			return nil
		},
	}
}
