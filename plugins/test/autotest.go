package test

import (
	"fmt"
	"log"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("test.autotest", TestAutotest{})
}

type TestAutotest struct{}

func (p TestAutotest) Run(data manifest.Manifest) error {
	if data.GetString("env") != data.GetString("current-env") {
		log.Printf("No test found for `%s`.\n", data.GetString("current-env"))
		return nil
	}

	if err := utils.RunCmd("rm -rf tests && git clone --depth 1 --single-branch --recursive %s tests", data.GetString("repo")); err != nil {
		return fmt.Errorf("Error on clone test git repo: %v", err)
	}

	return utils.RunCmd(
		"cd tests/ && ./test.sh --project=%s --version=%s --suite=%s",
		data.GetString("project"),
		data.GetString("version"),
		data.GetString("suite"),
	)
}
