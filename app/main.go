package app

import (
	"github.com/codegangsta/cli"
)

func AppCommand() cli.Command {
	return cli.Command{
		Name: "app",
		Subcommands: []cli.Command{
			BuildCommand(),
			DeployCommand(),
			ReleaseCommand(),
		},
	}
}
