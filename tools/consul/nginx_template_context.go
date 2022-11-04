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
	"github.com/servehub/utils/mergemap"
)

var upstreamNameRegex = regexp.MustCompile("[^\\w]+")
var spacesRegex = regexp.MustCompile("\\s+")
var allowedRouteVars = []string{"stage", "canary"}

func NginxTemplateContextCommand() cli.Command {
	return cli.Command{
		Name:  "nginx-template-context",
		Usage: "Collect and return data for consul-template",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "filter"},
			cli.StringFlag{Name: "exclude"},
		},
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			filters := parseFilters(c.String("filter"))
			excludes := parseFilters(c.String("exclude"))

			upstreams := make(map[string]map[string]map[string]interface{})
			services := make(map[string]map[string][]map[string]interface{})
			hostsParams := make(map[string]map[string]interface{})
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
					fmt.Fprintln(os.Stderr, color.RedString("Error on parse route json: %s. Service `%s`, json: %s", err, name, string(kv.Value)))
					continue
				}

				upstream := upstreamNameRegex.ReplaceAllString("serve_"+name, "_")
				if instances[0].Service.Port == 0 {
					upstream += "_static"
				}

				for _, route := range routes.Routes {
					for _, host := range spacesRegex.Split(route.Host, -1) {

						skipedByFilters := false
						for fk, fv := range filters {
							if vval, ok := route.Vars[fk]; !ok || vval != fv {
								skipedByFilters = true
								break
							}
						}

						if skipedByFilters {
							break
						}

						for ek, ev := range excludes {
							if vval, ok := route.Vars[ek]; ok || vval == ev {
								skipedByFilters = true
								break
							}
						}

						if skipedByFilters {
							break
						}

						for k, _ := range route.Vars {
							if !utils.Contains(k, allowedRouteVars) {
								delete(route.Vars, k)
							}
						}

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

						ssl := route.Ssl
						if _, ok := duplicates[host+":ssl"]; !ok {
							duplicates[host+":ssl"] = "exist"
						} else {
							ssl = nil
						}

						if route.Params == nil {
							route.Params = make(map[string]interface{})
						}

						if _, ok := hostsParams[host]; !ok {
							hostsParams[host] = make(map[string]interface{}, 0)
						}

						if res, err := mergemap.Merge(hostsParams[host], route.Params); err == nil {
							hostsParams[host] = res
						}

						services[host][location] = append(services[host][location], map[string]interface{}{
							"upstream":    upstream,
							"routeKeys":   routeKeys,
							"routeValues": routeValues,
							"sortIndex":   strconv.Itoa(len(route.Vars)),
							"cache":       route.Cache,
							"ssl":         ssl,
							"extra":       route.Extra,
							"params":      route.Params,
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
				"upstreams":   upstreams,
				"services":    services,
				"hostsParams": hostsParams,
			}, "", "  ")

			fmt.Fprintln(os.Stdout, string(out))
			return nil
		},
	}
}

func parseFilters(filter string) map[string]string {
	filters := make(map[string]string)
	for _, filter := range strings.Split(filter, ",") {
		if filter != "" {
			fvs := strings.SplitN(filter, "=", 2)

			if len(fvs) > 1 {
				filters[fvs[0]] = strings.TrimSpace(fvs[1])
			} else {
				filters[fvs[0]] = "true"
			}
		}
	}
	return filters
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
	Host     string                 `json:"host"`
	Location string                 `json:"location,omitempty"`
	Vars     map[string]string      `json:"vars,omitempty"`
	Cache    map[string]string      `json:"cache,omitempty"`
	Ssl      map[string]string      `json:"ssl,omitempty"`
	Extra    string                 `json:"extra,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
}
