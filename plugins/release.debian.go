package plugins

import (
	"fmt"

	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("release.debian", ReleaseDebian{})
}

type ReleaseDebian struct{}

func (p ReleaseDebian) Run(data manifest.Manifest) error {
	return runSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprintf(
			"sudo %s/debian-way/release.sh --package='%s' --site='%s' --mode='%s'",
			data.GetString("ci-tools-path"),
			data.GetString("package"),
			data.GetString("site"),
			data.GetString("mode"),
		),
		DefaultSshMaxProcs,
	)
}
