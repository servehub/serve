package build

import (
	"fmt"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("build.gradle", GradleBuild{})
}

type GradleBuild struct{}

func (p GradleBuild) Run(data manifest.Manifest) error {
	data.Set("cmd", fmt.Sprintf(data.GetString("cmd"), data.GetString("gradle")) )
	return manifest.PluginRegestry.Get("build.docker").Run(data)
}
