package test

import (
	"log"

	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("test", TestPlugin{"test"})
}

type TestPlugin struct {
	Name string
}

func (p TestPlugin) Run(data manifest.Manifest) error {
	log.Println("Run", p.Name, " --> ", data)
	return nil
}

