package build

import (
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.docker-run", BuildDockerRun{})
}

type BuildDockerRun struct{}

func (p BuildDockerRun) Run(data manifest.Manifest) error {
	envs := make([]string, 0)
	for k, v := range data.GetTree("envs").ToEnvMap("") {
		envs = append(envs, "-e " + k + "=" + v)
	}

	return utils.RunCmd(
		`docker run --rm %s -v "$PWD":/src -w /src %s /bin/sh -c '%s'`,
		strings.Join(envs, " "),
		data.GetString("image"),
		data.GetString("cmd"),
	)
}
