package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("upload.marathon", UploadMarathon{})
}

type UploadMarathon struct{}

func (p UploadMarathon) Run(data manifest.Manifest) error {
	if err := utils.RunCmd("curl -vsSf -o marathon.tar.gz %s", data.GetString("unstable-url")); err != nil {
		return err
	}

	return utils.RunCmd("curl -vsSf -XPUT -T marathon.tar.gz %s", data.GetString("stable-url"))
}
