package plugins

import (
	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("test", Test{})
}

type Test struct{}

func (p Test) Run(data manifest.Manifest) error {
	return nil
}
