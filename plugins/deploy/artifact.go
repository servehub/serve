package deploy

import (
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.artifact", DeployArtifact{})
}

type DeployArtifact struct{}

func (p DeployArtifact) Run(data manifest.Manifest) error {
	file := data.GetStringOr("artifact", "")

	if data.GetString("current-branch") != data.GetString("branch") || file == "" {
		return nil
	}

	return utils.RunCmd(
		"curl -sSf -u%s -T %s '%s%s/'",
		data.GetString("auth"),
		file,
		data.GetString("artifactory-url"),
		strings.TrimRight(data.GetString("upload-path"), "/"),
	)
}
