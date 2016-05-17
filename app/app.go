package app

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/app/deploy"
	"github.com/InnovaCo/serve/app/build"
)

var strategies = map[string]Strategy{
	"build.shell": build.ShellBuild{},
	"build.marathon": build.MarathonBuild{},
	"deploy.site": deploy.SiteDeploy{},
	"release.site": deploy.SiteRelease{},
}

func AppCommand() cli.Command {
	return cli.Command{
		Name: "app",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "env"},
			cli.StringFlag{Name: "feature"},
			cli.StringFlag{Name: "build-number",Value:"0"},
		},
		Subcommands: []cli.Command{
			BuildCommand(),
			DeployCommand(),
			ReleaseCommand(),
		},
	}
}

type Strategy interface {
	Run(m *manifest.Manifest, sub *manifest.Manifest) error
}

func GetStrategy(strategyType string, name string) (Strategy, error) {
	if m, ok := strategies[strategyType + "." + name]; ok {
		return m, nil
	} else {
		return nil, fmt.Errorf("Unknown deploy strategy `%v.%v`", strategyType, name)
	}
}
