package test

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("test.integration", TestIntegration{})
}

type TestIntegration struct{}

func (p TestIntegration) Run(data manifest.Manifest) error {
	return utils.RunCmd(data.GetString("command"))
}
