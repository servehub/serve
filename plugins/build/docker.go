package build

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.docker", BuildDockerRun{})
}

type BuildDockerRun struct{}

var nameRegepx = regexp.MustCompile("[^\\w.]+")

func (p BuildDockerRun) Run(data manifest.Manifest) error {
	envs := make([]string, 0)
	for k, v := range data.GetTree("envs").ToEnvMap("") {
		envs = append(envs, "-e "+k+"="+v)
	}

	cmds := make([]string, 0)
	for _, s := range data.GetArrayForce("cmd") {
		cmds = append(cmds, fmt.Sprintf("%s", s))
	}

	volumes := []string{fmt.Sprintf(`-v "%s":/src`, data.GetString("workdir"))}
	for _, v := range data.GetArrayForce("volumes") {
		volumes = append(volumes, fmt.Sprintf("-v %s", v))
	}

	image := data.GetString("image")

	if data.Has("build") {
		image += ":" + strings.ToLower(nameRegepx.ReplaceAllString(data.GetString("build"), ""))
		if err := utils.RunCmd("docker build --pull -t %s -f %s .", image, data.GetString("build")); err != nil {
			return err
		}
	} else {
		if err := utils.RunCmd("docker pull %s", image); err != nil {
			return err
		}
	}

	return utils.RunCmd(
		`docker run --rm %s %s -w /src %s %s %s`,
		strings.Join(envs, " "),
		strings.Join(volumes, " "),
		data.GetString("docker-args"),
		image,
		fmt.Sprintf(data.GetString("shell"), strings.Join(cmds, " && ")),
	)
}
