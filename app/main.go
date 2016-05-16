package app

import (
	"github.com/codegangsta/cli"
)

func AppCommand() cli.Command {
	return cli.Command{
		Name: "app",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "env"},
			cli.StringFlag{Name: "branch"},
			cli.StringFlag{Name: "build-number"},
		},
		Subcommands: []cli.Command{
			BuildCommand(),
			DeployCommand(),
			ReleaseCommand(),
		},
	}
}
