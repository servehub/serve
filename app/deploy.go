package app

import (
	"log"

	"github.com/codegangsta/cli"
)

func DeployCommand() cli.Command {
	return cli.Command{
		Name:  "deploy",
		Usage: "Deploy service",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "env"},
			cli.StringFlag{Name: "branch"},
			cli.StringFlag{Name: "build-number"},
		},
		Action: func(c *cli.Context) {
			log.Println("deploy")
		},
	}
}
