package build

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.marathon", MarathonBuild{})
}

type MarathonBuild struct{}

func (p MarathonBuild) Run(data manifest.Manifest) error {
	if err := utils.RunCmd("tar -zcf marathon.tar.gz -C %s/ .", data.GetString("source")); err != nil {
		return err
	}

	return utils.RunCmd("curl -vsSf -XPUT -T marathon.tar.gz %s", data.GetString("registry-url"))
}
