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
	err := utils.RunCmd(
		`dig +short %s | sort | uniq | parallel -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no %s@{} "sudo %s/debian-way/deploy.sh --package=%s"`,
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		data.GetString("ci-tools-path"),
		data.GetString("package"),
	)
	if err != nil {
		return err
	}

	consulApi, err := ConsulClient(data.GetString("consul-host"))
	if err != nil {
		return err
	}

	if err := setKey(consulApi, "/plugins/" + data.GetString("package") + "/deploy.debian", data.String()); err != nil {
		return err
	}

	return nil
}
