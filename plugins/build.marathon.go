package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.marathon", MarathonBuild{})
}

type MarathonBuild struct{}

func (p MarathonBuild) Run(data manifest.Manifest) error {
	if err := utils.RunCmdf("tar -zcf marathon.tar.gz -C %s/ .", data.GetString("source")); err != nil {
		return err
	}

	return utils.RunCmdf("curl -vsSf -XPUT -T marathon.tar.gz %s", data.GetString("registry-url"))
}
