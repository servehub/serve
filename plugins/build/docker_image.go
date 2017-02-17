package build

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.docker-image", BuildDockerImage{})
}

type BuildDockerImage struct{}

func (p BuildDockerImage) Run(data manifest.Manifest) error {
	if err := utils.RunCmd("docker build --pull -t %s .", data.GetString("image")); err != nil {
		return err
	}

	return utils.RunCmd("docker push %s", data.GetString("image"))
}
