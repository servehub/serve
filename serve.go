package main

import (
	"os"

	"github.com/codegangsta/cli"

	appCmd "github.com/kulikov/serve/app"
	"github.com/kulikov/serve/consul"
	"github.com/kulikov/serve/github"
	"github.com/kulikov/serve/marathon"
)

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Version = "0.2"
	app.Usage = "Automate your infrastructure!"

	app.Commands = []cli.Command{
		appCmd.AppCommand(),
		consul.ConsulCommand(),
		marathon.MarathonCommand(),
		github.WebhookServerCommand(),
	}

	app.Run(os.Args)
}
