package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"log"
	"github.com/fatih/color"
	"fmt"
	"github.com/InnovaCo/serve/utils"
)

const goUser = "go-agent"
const ciToolsPath = "/local/innova/tools"

func init() {
	manifest.PluginRegestry.Add("deploy.debian", DeployDebian{})
}

type DeployDebian struct {}

func(p DeployDebian) Run(data manifest.Manifest, vars map[string]string) error {
	log.Println(color.GreenString("Start deploy.debian plugin"))
	nameVersion := data.GetString("name-version")
	if (nameVersion == "") {
		log.Fatal(color.RedString("`name-version` is not defined in manifest!"))
	}

	cluster := data.GetString("cluster")
	if (cluster == "") {
		log.Fatal(color.RedString("`cluster` is not defined in manifest!"))
	}

	return utils.RunCmd(
		fmt.Sprintf(
			"dig +short %s | sort | uniq | parallel -j 10 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no %s" +
				"\"sudo %s/inn-ci-tools/debian-way/deploy.sh --package=%s\"",
			cluster, goUser, ciToolsPath, nameVersion))
}
