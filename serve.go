package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"

	appCmd "github.com/InnovaCo/serve/app"
	"github.com/InnovaCo/serve/consul"
	"github.com/InnovaCo/serve/marathon"
)

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Version = "0.3"
	app.Usage = "Automate your infrastructure!"

	app.Commands = []cli.Command{
		appCmd.AppCommand(),
		consul.ConsulCommand(),
		marathon.MarathonCommand(),
		//github.WebhookServerCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(color.RedString("Exit: %v", err))
	}
}
