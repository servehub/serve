package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("upload.debian", UploadDebian{})
}

type UploadDebian struct{}

func (p UploadDebian) Run(data manifest.Manifest) error {
	return utils.RunCmd(
		`ssh %s "%s %s %s %s"`,
		data.GetString("ssh-connection"),
		data.GetString("script"),
		data.GetString("changes-file"),
		data.GetString("src-repo"),
		data.GetString("dst-repo"),
	)
}
