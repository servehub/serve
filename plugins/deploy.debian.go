package plugins

import (
	"fmt"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.debian", DeployDebian{})
}

const SshMaxProcs=1

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
		data.GetIntOr("parallel", SshMaxProcs),

	); err != nil {
		return err
	}

	return registerPluginData("deploy.debian", data.GetString("app-name"), data.String(), data.GetString("consul-host"))
}

func (p DeployDebian) Uninstall(data manifest.Manifest) error {
	if err := runSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo apt-get purge %s -y", data.GetString("package")),
		1,
	); err != nil {
		return err
	}

	return deletePluginData("deploy.debian", data.GetString("app-name"), data.GetString("consul-host"))
}

func runSshCmd(cluster, sshUser, cmd string, maxProcs int) error {
	sshCmd := "ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
	if maxProcs > 1 {
		if maxProcs > SshMaxProcs {
			maxProcs = SshMaxProcs
		}
		return utils.RunCmd(
			`dig +short %s | sort | uniq | parallel --tag --line-buffer -j %d %s %s@{} "%s"`,
			cluster,
			maxProcs,
			sshCmd,
			sshUser,
			cmd,
		)
	} else {
		return utils.RunCmd(
			`%s %s@{%s} "%s"`,
			sshCmd,
			sshUser,
			cluster,
			cmd,
		)
	}
}