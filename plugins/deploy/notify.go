package deploy

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/plugins/build"
)

func init() {
	manifest.PluginRegestry.Add("deploy.notify", build.BuildNotify{})
}
