package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"

	"github.com/InnovaCo/serve/tools/consul"
	"github.com/InnovaCo/serve/tools/supervisor"
	"github.com/InnovaCo/serve/tools/testrunner"
)

var version = "0.0"

func main() {
	app := cli.NewApp()
	app.Name = "serve-tools"
	app.Version = version
	app.Usage = "Serve tools"

	app.Commands = []cli.Command{
		cli.Command{
			Name: "consul",
			Subcommands: []cli.Command{
				consul.SupervisorCommand(),
				consul.NginxTemplateContextCommand(),
				consul.RouteCommand(),
			},
		},
		supervisor.SupervisorCommand(),
		testrunner.TestRunnerCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(color.RedString("Exit: %v", err))
	}
}
