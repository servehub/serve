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
	image := data.GetString("image")
	prefix := image[:strings.Index(image, ":")]

	if data.Has("category") {
		prefix = prefix[:strings.Index(prefix, "/")] + "/" + data.GetString("category") + prefix[strings.LastIndex(prefix, "/"):]
	}

	if data.Has("name") {
		prefix = prefix[:strings.LastIndex(prefix, "/")] + "/" + data.GetString("name")
	}

	if data.Has("login.user") {
		if err := utils.RunCmd(
			`docker login -u "%s" -p "%s" %s`,
			data.GetString("login.user"),
			data.GetString("login.password"),
			image[:strings.Index(image, "/")],
		); err != nil {
			return err
		}
	}

	tags := make([]string, 0)
	for _, tag := range data.GetArrayForce("tags") {
		tags = append(tags, fmt.Sprintf("%s:%v", prefix, tag))
	}

	// pull exists tagged images for cache
	for _, tag := range tags {
		utils.RunCmd("docker pull %s", tag)
	}

	cacheFrom := ""
	if len(tags) == 0 {
		tags = []string{image}
	} else {
		cacheFrom = "--cache-from=" + tags[0]
	}

	if err := utils.RunCmd(
		"docker build --pull -t %s %s %s",
		strings.Join(tags, " -t "),
		cacheFrom,
		data.GetString("workdir"),
	); err != nil {
		return err
	}

	if !data.GetBool("no-push") {
		for _, tag := range tags {
			if err := utils.RunCmd("docker push %s", tag); err != nil {
				return err
			}
		}
	}

	return nil
}
