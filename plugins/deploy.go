package app

import (
	"log"

	"github.com/InnovaCo/serve/manifest"
	"github.com/codegangsta/cli"
)

func DeployCommand() cli.Command {
	return cli.Command{
		Name:  "deploy",
		Usage: "Deploy service",
		Action: func(c *cli.Context) error {
			mf := manifest.LoadManifest(c)

			if mf.Has("deploy") {
				strategy, err := GetStrategy("deploy", mf.GetStringOr("deploy.type", "default"))

				if err != nil {
					log.Fatalf("Deploy error: %v", err)
				}

				return strategy.Run(mf, mf.Sub("deploy"))
			}

			return nil
		},
	}
}
