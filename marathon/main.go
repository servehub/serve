package marathon

import (
	"github.com/codegangsta/cli"
)

func MarathonCommand() cli.Command {
	return cli.Command{
		Name: "marathon",
		Subcommands: []cli.Command{
			DeploySiteCommand(),
			ReleaseSiteCommand(),
			DeployTaskCommand(),
		},
	}
}
