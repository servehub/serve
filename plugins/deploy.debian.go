package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
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
	if err := utils.RunCmd(
		`dig +short %s | sort | uniq | parallel -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no %s@{} "sudo %s/debian-way/deploy.sh --package=%s"`,
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		data.GetString("ci-tools-path"),
		data.GetString("package"),
	); err != nil {
		return err
	}

	return registerPluginData("deploy.debian", data.GetString("package"), data.String(), data.GetString("consul-host"))
}

func (p DeployDebian) Uninstall(data manifest.Manifest) error {
	if err := utils.RunCmd(
		`dig +short %s | sort | uniq | parallel -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no %s@{} "sudo apt-get purge %s"`,
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		data.GetString("ci-tools-path"),
		data.GetString("package"),
	); err != nil {
		return err
	}

	return deletePluginData("deploy.debian", data.GetString("package"), data.GetString("consul-host"))
}
