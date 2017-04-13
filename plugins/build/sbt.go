package build

import (
	"fmt"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("build.sbt", SbtBuild{})
}

type SbtBuild struct{}

func (p SbtBuild) Run(data manifest.Manifest) error {
	data.Set("cmd", fmt.Sprintf(data.GetString("cmd"), data.GetString("sbt")) )
	return manifest.PluginRegestry.Get("build.docker").Run(data)
}
