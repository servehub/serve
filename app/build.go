package app

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"

	"github.com/InnovaCo/serve/manifest"
)

func BuildCommand() cli.Command {
	return cli.Command{
		Name:  "build",
		Usage: "Build package",
		Action: func(c *cli.Context) error {

			mf := manifest.LoadManifest(c)

			for _, bldr := range mf.Array("build") {
				name, err := bldr.FirstKey()
				if err != nil {
					panic(color.RedString("Build error: %v", err))
				}

				strategy, err := GetStrategy("build", name)
				if err != nil {
					panic(color.RedString("Build error: %v", err))
				}

				log.Printf("build %s -> %v", name, bldr)
				if err := strategy.Run(mf, bldr); err != nil {
					return err
				}

				println("")
			}

			return nil
		},
	}
}
