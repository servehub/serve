package main

import (
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"

	appCmd "github.com/InnovaCo/serve/app"
	"github.com/InnovaCo/serve/consul"
)

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Version = "0.3"
	app.Usage = "Automate your infrastructure!"

	app.Commands = []cli.Command{
		appCmd.AppCommand(),
		consul.ConsulCommand(),
		//github.WebhookServerCommand(),
	}

	app.Before = func(c *cli.Context) error {
		log.Println(color.MagentaString(">>> %s\n", strings.Join(c.Args(), " ")))
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(color.RedString("Exit: %v", err))
	}
}
