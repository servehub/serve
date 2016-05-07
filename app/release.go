package app

import (
	"log"

	"github.com/codegangsta/cli"
)

func ReleaseCommand() cli.Command {
	return cli.Command{
		Name:  "release",
		Usage: "Release service",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "env"},
			cli.StringFlag{Name: "branch"},
			cli.StringFlag{Name: "build-number"},
		},
		Action: func(c *cli.Context) {
			log.Println("release")
		},
	}
}
