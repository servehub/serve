package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.sbt_pack", SbtPackBuild{})
}

type SbtPackBuild struct{}

func (p SbtPackBuild) Run(data manifest.Manifest, vars map[string]string) error {
	if err := utils.RunCmdf("sbt ';set version := \"%s\"' clean test pack", data.GetString("version")); err != nil {
		return err
	}

	return nil
}
