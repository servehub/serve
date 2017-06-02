package build

import (
	"fmt"
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.docker-image", BuildDockerImage{})
}

type BuildDockerImage struct{}

func (p BuildDockerImage) Run(data manifest.Manifest) error {
	if data.Has("login.user") {
		if err := utils.RunCmd(
			`docker login -u "%s" -p "%s" %s`,
			data.GetString("login.user"),
			data.GetString("login.password"),
			data.GetString("login.registry"),
		); err != nil {
			return err
		}
	}

	image := data.GetString("image")
	prefix := image[:strings.Index(image, ":")]

	if data.Has("name") {
		prefix = prefix[:strings.LastIndex(prefix, "/")] + "/" + data.GetString("name")
	}

	tags := make([]string, 0)

	for _, tag := range data.GetArray("tags") {
		tags = append(tags, fmt.Sprintf("%s:%v", prefix, tag.Unwrap()))
	}

	// pull exists tagged images for cache
	for _, tag := range tags {
		utils.RunCmd("docker pull %s", tag)
	}

	if len(tags) == 0 {
		tags = []string{image}
	}

	if err := utils.RunCmd(
		"docker build --pull -t %s %s",
		strings.Join(tags, " -t "),
		data.GetString("workdir"),
	); err != nil {
		return err
	}

	for _, tag := range tags {
		if err := utils.RunCmd("docker push %s", tag); err != nil {
			return err
		}
	}

	return nil
}
