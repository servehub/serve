package consul

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"
)

func NginxTemplateTcpContextCommand() cli.Command {
	return cli.Command{
		Name:  "nginx-template-tcp-context",
		Usage: "Collect and return data for consul-template for TCP/UDP streams",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "filter"},
		},
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			filters := make(map[string]string)
			for _, filter := range strings.Split(c.String("filter"), ",") {
				if filter != "" {
					fvs := strings.SplitN(filter, "=", 2)

					if len(fvs) > 1 {
						filters[fvs[0]] = strings.TrimSpace(fvs[1])
					} else {
						filters[fvs[0]] = "true"
					}
				}
			}

			upstreams := make(map[string]map[string]map[string]interface{})
			services := make(map[int]map[string]interface{})

			allServicesRoutes, _, err := consul.KV().List("services/tcp-routes/", nil)
			if err != nil {
				return fmt.Errorf("Error on load routes: %s", err)
			}

			for _, kv := range allServicesRoutes {
				name := strings.TrimPrefix(kv.Key, "services/tcp-routes/")

				instances, _, err := consul.Health().Service(name, "", true, nil)
				if err != nil {
					return fmt.Errorf("Error on get service `%s` health: %s", name, err)
				}

				if len(instances) == 0 {
					continue
				}

				tcp := consulTcpRoute{}
				if err := json.Unmarshal(kv.Value, &tcp); err != nil {
					return fmt.Errorf("Error on parse tcp json: %s. Serive `%s`, json: %s", err, name, string(kv.Value))
				}

				upstream := upstreamNameRegex.ReplaceAllString("serve_tcp_"+name, "_")

				skipedByFilters := false

				for fk, fv := range filters {
					if vval, ok := tcp.Vars[fk]; !ok || vval != fv {
						skipedByFilters = true
						break
					}

					delete(tcp.Vars, fk)
				}

				if skipedByFilters {
					continue
				}

				delete(tcp.Vars, "public") // todo: remove hardcoded filter

				for _, inst := range instances {
					putUpstream(upstream, inst, upstreams)
				}

				services[tcp.Port] = map[string]interface{}{
					"upstream": upstream,
					"protocol": "",
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

type consulTcpRoute struct {
	Port int               `json:"port"`
	Vars map[string]string `json:"vars,omitempty"`
}
