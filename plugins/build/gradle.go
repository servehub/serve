package build

import (
	"fmt"
	"strings"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("build.gradle", GradleBuild{})
}

type GradleBuild struct{}

func (p GradleBuild) Run(data manifest.Manifest) error {
	cmd := fmt.Sprintf(data.GetString("cmd"), data.GetString("gradle"))

	if data.GetBool("no-push") {
		cmd = strings.ReplaceAll(cmd, " publish", "")
	}

	data.Set("cmd", cmd)

	return manifest.PluginRegestry.Get("build.docker").Run(data)
}
