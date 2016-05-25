package app

import (
	"fmt"

	"github.com/InnovaCo/serve/app/build"
	"github.com/InnovaCo/serve/app/deploy/site"
	"github.com/InnovaCo/serve/manifest"
	"github.com/codegangsta/cli"
)

var strategies = map[string]Strategy{
	"build.shell":    build.ShellBuild{},
	"build.sbt-pack": build.SbtPackBuild{},
	"build.marathon": build.MarathonBuild{},
	"deploy.site":    site.SiteDeploy{},
	"release.site":   site.SiteRelease{},
}

func AppCommand() cli.Command {
	commonFlags := []cli.Flag{
		cli.StringFlag{Name: "env"},
		cli.StringFlag{Name: "feature"},
		cli.StringFlag{Name: "build-number", Value: "0"},
		cli.StringFlag{Name: "manifest", Value: "manifest.yml"},
	}

	return cli.Command{
		Name:  "app",
		Subcommands: []cli.Command{
			withFlags(BuildCommand(), commonFlags),
			withFlags(DeployCommand(), commonFlags),
			withFlags(ReleaseCommand(), commonFlags),
		},
	}
}

type Strategy interface {
	Run(m *manifest.Manifest, sub *manifest.Manifest) error
}

func GetStrategy(strategyType string, name string) (Strategy, error) {
	if m, ok := strategies[strategyType+"."+name]; ok {
		return m, nil
	} else {
		return nil, fmt.Errorf("Unknown deploy strategy `%v.%v`", strategyType, name)
	}
}

func withFlags(cmd cli.Command, commonFlags []cli.Flag) cli.Command {
	cmd.Flags = append(commonFlags, cmd.Flags...)
	return cmd
}
