package plugins

import (
	"fmt"

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
	if err := runSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo %s/debian-way/deploy.sh --package='%s' --version='%s'", data.GetString("ci-tools-path"), data.GetString("package"), data.GetString("version")),
	); err != nil {
		return err
	}

	return registerPluginData("deploy.debian", data.GetString("package"), data.String(), data.GetString("consul-host"))
}

func (p DeployDebian) Uninstall(data manifest.Manifest) error {
	if err := runSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo apt-get -y purge %s", data.GetString("package")),
	); err != nil {
		return err
	}

	return deletePluginData("deploy.debian", data.GetString("package"), data.GetString("consul-host"))
}

func runSshCmd(cluster, sshUser, cmd string) error {
	return utils.RunCmd(
		`dig +short %s | sort | uniq | parallel -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no %s@{} "%s"`,
		cluster,
		sshUser,
		cmd,
	)
}
