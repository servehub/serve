package build

import (
	"fmt"
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.docker", BuildDockerRun{})
}

type BuildDockerRun struct{}

func (p BuildDockerRun) Run(data manifest.Manifest) error {
	envs := make([]string, 0)
	for k, v := range data.GetTree("envs").ToEnvMap("") {
		envs = append(envs, "-e "+k+"="+v)
	}

	return utils.RunCmd(
		`docker run --rm %s -v "$PWD":/src -w /src %s %s`,
		strings.Join(envs, " "),
		data.GetString("image"),
		fmt.Sprintf(data.GetString("shell"), data.GetString("cmd")),
	)
}
