package plugins

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("command", Command{})
}

type Command struct{}

func (p Command) Run(data manifest.Manifest) error {
	return utils.RunCmd(data.GetString("command"))
}
