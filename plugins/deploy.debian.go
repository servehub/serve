package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"log"
	"github.com/fatih/color"
	"fmt"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.debian", DeployDebian{})
}

type DeployDebian struct {}

func(p DeployDebian) Run(data manifest.Manifest) error {
	nameVersion := data.GetString("name-version")
	if (nameVersion == "") {
		log.Fatal(color.RedString("`name-version` is not defined in manifest!"))
	}

	cluster := data.GetString("cluster")
	if (cluster == "") {
		log.Fatal(color.RedString("`cluster` is not defined in manifest!"))
	}

	// take gouser , toolsPath from data
	sshUser := data.GetString("ssh-user")
	ciToolsPath := data.GetString("ci-tools-path")
	if (ciToolsPath == "") {
		log.Fatal(color.RedString("`ci-tools-path` is not defined in manifest!"))
	}

	return utils.RunCmd(
		fmt.Sprintf(
			"dig +short %s | sort | uniq | parallel -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no %s" +
				"\"sudo %s/debian-way/deploy.sh --package=%s\"",
			cluster, sshUser, ciToolsPath, nameVersion))
}
