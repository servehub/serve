package build

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.sbt", SbtBuild{})
}

type SbtBuild struct{}

func (p SbtBuild) Run(data manifest.Manifest) error {
  return utils.RunCmd(`sbt ';set version := "%s"' clean test %s`, data.GetString("version"), data.GetStringOr("sbt", ""))
}
