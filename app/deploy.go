package app

import (
	"log"

	"github.com/codegangsta/cli"
)

func DeployCommand() cli.Command {
	return cli.Command{
		Name:  "deploy",
		Usage: "Deploy service",
		Action: func(c *cli.Context) {
			// запускаем новую версию сервиса
			// дожидаемся появления в консуле

			log.Println("deploy")
		},
	}
}
