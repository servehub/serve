package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"

	appCmd "github.com/InnovaCo/serve/app"
	"github.com/InnovaCo/serve/consul"
	"github.com/InnovaCo/serve/supervisor"
)

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Version = "0.3"
	app.Usage = "Automate your infrastructure!"

	app.Commands = []cli.Command{
		appCmd.AppCommand(),
		consul.ConsulCommand(),
		supervisor.SupervisorCommand(),
		//github.WebhookServerCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(color.RedString("Exit: %v", err))
	}
}
