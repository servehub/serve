package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"

	"github.com/servehub/serve/tools/consul"
	"github.com/servehub/serve/tools/supervisor"
	"github.com/servehub/serve/tools/testrunner"
)

var version = "0.0"

func main() {
	app := cli.NewApp()
	app.Name = "serve-tools"
	app.Version = version
	app.Usage = "Serve tools"

	app.Commands = []cli.Command{
		{
			Name: "consul",
			Subcommands: []cli.Command{
				consul.SupervisorCommand(),
				consul.NginxTemplateContextCommand(),
				consul.NginxTemplateTcpContextCommand(),
				consul.KvPatchCommand(),
				consul.DeregisterCommand(),
				{
					Name: "kv",
					Subcommands: []cli.Command{
						consul.KvPatchCommand(),
						consul.KvRenameCommand(),
					},
				},
			},
		},
		supervisor.SupervisorCommand(),
		testrunner.TestRunnerCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(color.RedString("Exit: %v", err))
	}
}
