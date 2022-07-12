package release

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("release.sh", ShRelease{})
}

type ShRelease struct{}

func (p ShRelease) Run(data manifest.Manifest) error {
	return utils.RunCmd(data.GetString("sh"))
}
