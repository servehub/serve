package build

import (
	"fmt"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("build.node", NodeBuild{})
}

type NodeBuild struct{}

func (p NodeBuild) Run(data manifest.Manifest) error {
	data.Set("cmd", fmt.Sprintf(data.GetString("cmd"), data.GetString("node")))
	return manifest.PluginRegestry.Get("build.docker").Run(data)
}
