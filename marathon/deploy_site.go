package marathon

import (
	"log"

	"github.com/codegangsta/cli"
	marathon "github.com/gambol99/go-marathon"
)

func DeploySiteCommand() cli.Command {
	return cli.Command{
		Name:  "deploy-site",
		Usage: "Deploy site into marathon",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "marathon"},
			cli.StringFlag{Name: "name"},
			cli.StringFlag{Name: "version"},
			cli.StringFlag{Name: "host"},
			cli.StringFlag{Name: "location"},
			cli.StringFlag{Name: "staging"},
			cli.IntFlag{Name: "instances"},
			cli.IntFlag{Name: "mem"},
			cli.StringFlag{Name: "constraints"},
			cli.StringFlag{Name: "envs"},
		},
		Action: func(c *cli.Context) {
			marathonConf := marathon.NewDefaultConfig()
			marathonConf.URL = "http://" + c.GlobalString("marathon-host") + ":8080"
			marathonApi, _ := marathon.NewClient(marathonConf)

			log.Println(marathonApi)

			//app := &marathon.Application{
			//	ID: c.GlobalString("name") + "-v" + c.GlobalString("version"),
			//	Cmd: "bin/serve service --name $(echo '#{project}' | sed 's/[^a-z0-9]/-/gI') --version #{version}.${GO_PIPELINE_LABEL} --qa-domain '#{domain}' --location '#{location}' --port \\$PORT0",
			//}
			//
			//marathonApi.UpdateApplication(app, false)
		},
	}
}
