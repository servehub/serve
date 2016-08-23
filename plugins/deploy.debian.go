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
	if err := runParallelSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo %s/debian-way/deploy.sh --package='%s' --version='%s'", data.GetString("ci-tools-path"), data.GetString("package"), data.GetString("version")),
		data.GetIntOr("parallel", 1),
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
	); err != nil {
		return err
	}

	return deletePluginData("deploy.debian", data.GetString("app-name"), data.GetString("consul-host"))
}

func runSshCmd(cluster, sshUser, cmd string) error {
	return runParallelSshCmd(cluster, sshUser, cmd, 1)
}

func runParallelSshCmd(cluster, sshUser, cmd string, maxProcs int) error {
	return utils.RunCmd(
		`dig +short %s | sort | uniq | parallel --tag --line-buffer -j %d ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null %s@{} "%s"`,
		cluster,
		maxProcs,
		sshUser,
		cmd,
	)
}
