package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.sbt-pack", SbtPackBuild{})
}

type SbtPackBuild struct{}

func (p SbtPackBuild) Run(data manifest.Manifest) error {
	return utils.RunCmd(`sbt ';set version := "%s"' clean test pack`, data.GetString("version"))
}
