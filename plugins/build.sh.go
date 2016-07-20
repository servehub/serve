package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.sh", ShBuild{})
}

type ShBuild struct{}

func (p ShBuild) Run(data manifest.Manifest, vars map[string]string) error {
	return utils.RunCmd(data.GetString("sh"))
}
