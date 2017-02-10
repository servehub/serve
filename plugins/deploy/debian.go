package deploy

import (
	"fmt"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.debian", DeployDebian{})
}

type DeployDebian struct{}

func (p DeployDebian) Run(data manifest.Manifest) error {
	if data.GetBool("purge") {
		return p.Uninstall(data)
	} else {
		return p.Install(data)
	}
}

func (p DeployDebian) Install(data manifest.Manifest) error {
	if err := utils.RunParallelSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo %s/debian-way/deploy.sh --package='%s' --version='%s'", data.GetString("ci-tools-path"), data.GetString("package"), data.GetString("version")),
		data.GetIntOr("parallel", 1),
	); err != nil {
		return err
	}

	return utils.RegisterPluginData("deploy.debian", data.GetString("app-name"), data.String(), data.GetString("consul-address"))
}

func (p DeployDebian) Uninstall(data manifest.Manifest) error {
	if err := utils.RunSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo apt-get purge %s -y", data.GetString("package")),
	); err != nil {
		return err
	}

	return utils.DeletePluginData("deploy.debian", data.GetString("app-name"), data.GetString("consul-address"))
}
