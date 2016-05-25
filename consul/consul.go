package consul

import (
	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/manifest"
)

func ConsulCommand() cli.Command {
	return cli.Command{
		Name: "consul",
		Subcommands: []cli.Command{
			SupervisorCommand(),
			NginxTemplateContextCommand(),
		},
	}
}

func ConsulClient(m *manifest.Manifest) *api.Client {
	conf := api.DefaultConfig()
	conf.Address = m.GetString("consul.consul-host") + ":8500"

	consul, _ := api.NewClient(conf)
	return consul
}
