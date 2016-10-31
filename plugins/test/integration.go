package test

import (
	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("test.integration", TestIntegration{})
}

type TestIntegration struct{}

func (p TestIntegration) Run(data manifest.Manifest) error {
	return nil
}
