package supervisor

import (
	"log"

	"github.com/codegangsta/cli"
)

func SupervisorCommand() cli.Command {
	return cli.Command{
		Name: "supervisor",
		Flags: []cli.Flag{
			cli.IntFlag{Name: "max-retry", Value: -1},
		},
		Subcommands: []cli.Command{
			{
				Name:            "start",
				SkipFlagParsing: true,
				Action: func(c *cli.Context) {
					log.Println("Starting", c.Args())

					super := Supervisor{
						MaxRetry: c.Int("max-retry"),
						OnError: func() {
							log.Println("Error")
						},
						OnComplete: func() {
							log.Println("Complete")
						},
					}

					super.Run(c.Args())
				},
			},
		},
	}
}
