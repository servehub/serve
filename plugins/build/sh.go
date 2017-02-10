package build

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.sh", ShBuild{})
}

type ShBuild struct{}

func (p ShBuild) Run(data manifest.Manifest) error {
	return utils.RunCmd(data.GetString("sh"))
}
