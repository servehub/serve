package app

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/InnovaCo/serve/manifest"
)

func ReleaseCommand() cli.Command {
	return cli.Command{
		Name:  "release",
		Usage: "Release service",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "env"},
			cli.StringFlag{Name: "feature"},
			cli.StringFlag{Name: "build-number"},
			cli.StringFlag{Name: "route"},
		},
		Action: func(c *cli.Context) error {
			mf := manifest.LoadManifest(c)

			if mf.Has("deploy") {
				strategy, err := GetStrategy("release", mf.GetStringOr("deploy.type", "default"))

				if err != nil {
					log.Fatalf("Release error: %v", err)
				}

				return strategy.Run(mf, mf.Sub("deploy"))
			}

			return nil
		},
	}
}
