package release

import (
	"fmt"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("release.debian", ReleaseDebian{})
}

type ReleaseDebian struct{}

func (p ReleaseDebian) Run(data manifest.Manifest) error {
	return utils.RunSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf(
			"sudo %s/debian-way/release.sh --package='%s' --site='%s' --mode='%s'",
			data.GetString("ci-tools-path"),
			data.GetString("package"),
			data.GetString("site"),
			data.GetString("mode"),
		),
	)
}
