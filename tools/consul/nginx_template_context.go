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

	"github.com/servehub/utils"
)

var upstreamNameRegex = regexp.MustCompile("[^\\w]+")
var spacesRegex = regexp.MustCompile("\\s+")

func NginxTemplateContextCommand() cli.Command {
	return cli.Command{
		Name:  "nginx-template-context",
		Usage: "Collect and return data for consul-template",
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			upstreams := make(map[string]map[string]map[string]interface{})
			services := make(map[string]map[string][]map[string]interface{})
			duplicates := make(map[string]string)

			allServicesRoutes, _, err := consul.KV().List("services/routes/", nil)
			if err != nil {
				return fmt.Errorf("Error on load routes: %s", err)
			}

			for _, kv := range allServicesRoutes {
				name := strings.TrimPrefix(kv.Key, "services/routes/")

				instances, _, err := consul.Health().Service(name, "", true, nil)
				if err != nil {
					return fmt.Errorf("Error on get service `%s` health: %s", name, err)
				}

				if len(instances) == 0 {
					continue
				}

				routes := consulRoutes{}
				if err := json.Unmarshal(kv.Value, &routes); err != nil {
					return fmt.Errorf("Error on parse route json: %s. Serive `%s`, json: %s", err, name, string(kv.Value))
				}

				upstream := upstreamNameRegex.ReplaceAllString("serve_"+name, "_")
				if instances[0].Service.Port == 0 {
					upstream += "_static"
				}

				for _, route := range routes.Routes {
					for _, host := range spacesRegex.Split(route.Host, -1) {
						location := route.Location
						if location == "" {
							location = "/"
						}

						for _, inst := range instances {
							putUpstream(upstream, inst, upstreams)
						}

						if _, ok := services[host]; !ok {
							services[host] = make(map[string][]map[string]interface{}, 0)
						}

						if _, ok := services[host][location]; !ok {
							services[host][location] = make([]map[string]interface{}, 0)
						}

						routeKeys := "-"
						routeValues := "-"
						for k, v := range route.Vars {
							routeKeys += "${" + k + "}-"
							routeValues += v + "-"
						}

						if exists, ok := duplicates[host+location+routeKeys+routeValues]; !ok {
							duplicates[host+location+routeKeys+routeValues] = name
						} else {
							fmt.Fprintln(os.Stderr, color.RedString("Service with the same routes already exists! exists: %s, skipped: %s", exists, name))
							continue
						}

						services[host][location] = append(services[host][location], map[string]interface{}{
							"upstream":    upstream,
							"routeKeys":   routeKeys,
							"routeValues": routeValues,
							"sortIndex":   strconv.Itoa(len(route.Vars)),
							"cache":       route.Cache,
						})
					}
				}
			}

			// sort routes by sort index
			for _, hh := range services {
				for _, ll := range hh {
					sort.Sort(utils.BySortIndex(ll))
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

func putUpstream(upstream string, inst *api.ServiceEntry, upstreams map[string]map[string]map[string]interface{}) {
	port := inst.Service.Port

	if _, ok := upstreams[upstream]; !ok {
		upstreams[upstream] = make(map[string]map[string]interface{}, 0)
	}

	address := inst.Node.Address
	if inst.Service.Address != "" {
		address = inst.Service.Address
	}

	upstreams[upstream][fmt.Sprintf("%s:%d", address, port)] = map[string]interface{}{
		"address": address,
		"port":    port,
	}
}

type consulRoutes struct {
	Routes []consulRoute `json:"routes"`
}

type consulRoute struct {
	Host     string            `json:"host"`
	Location string            `json:"location,omitempty"`
	Vars     map[string]string `json:"vars,omitempty"`
	Cache    map[string]string `json:"cache,omitempty"`
}
