package test

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("test.sh", ShTest{})
}

type ShTest struct{}

func (p ShTest) Run(data manifest.Manifest) error {
	return utils.RunCmd(data.GetString("sh"))
}
