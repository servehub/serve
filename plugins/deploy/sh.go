package deploy

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.sh", DeploySh{})
}

type DeploySh struct{}

func (p DeploySh) Run(data manifest.Manifest) error {
	return utils.RunCmd(data.GetString("sh"))
}
