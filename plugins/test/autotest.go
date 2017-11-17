package test

import (
	"fmt"
	"log"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("test.autotest", TestAutotest{})
}

type TestAutotest struct{}

func (p TestAutotest) Run(data manifest.Manifest) error {
	if data.GetString("env") != data.GetString("current-env") {
		log.Printf("No autotest found for `%s`.\n", data.GetString("current-env"))
		return nil
	}

	if err := utils.RunCmd("rm -rf autotest && git clone --depth 1 --single-branch --recursive %s autotest", data.GetString("repo")); err != nil {
		return fmt.Errorf("error on clone test git repo: %v", err)
	}

	envs := make(map[string]string, 0)
	for k, v := range data.GetMap("environment") {
		envs[k] = fmt.Sprintf("%s", v.Unwrap())
	}

	return utils.RunCmdWithEnv(fmt.Sprintf(
			"cd autotest/ && ./test.sh --project=%s --version=%s --suite=%s",
			data.GetString("project"),
			data.GetString("version"),
			data.GetString("suite"),
		),
		envs,
	)
}
