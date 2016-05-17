package app

import (
	"log"

	"github.com/InnovaCo/serve/manifest"
	"github.com/codegangsta/cli"
)

func BuildCommand() cli.Command {
	return cli.Command{
		Name:  "build",
		Usage: "Build package",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "branch"},
			cli.StringFlag{Name: "build-number"},
		},
		Action: func(c *cli.Context) error {

			mf := manifest.LoadManifest(c)

			for _, bldr := range mf.Array("build") {
				name, err := bldr.FirstKey()
				if err != nil {
					log.Fatalf("Build error: %v", err)
				}

				strategy, err := GetStrategy("build", name)
				if err != nil {
					log.Fatalf("Build error: %v", err)
				}

				if err := strategy.Run(mf, bldr); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

//type Manifest struct {
//	conf *viper.Viper
//}
//
//func (m *Manifest) GetString(key string) string {
//	return m.conf.GetString(key)
//}
