package build

import (
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"regexp"
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

	if data.Has("repository") {
		image = data.GetString("repository") + image[strings.Index(image, "/"):]
	}

	if data.Has("category") {
		image = image[:strings.Index(image, "/")] + "/" + data.GetString("category") + image[strings.LastIndex(image, "/"):]
	}

	if data.Has("name") {
		image = image[:strings.LastIndex(image, "/")] + "/" + data.GetString("name") + image[strings.Index(image, ":"):]
	}

	prefix := image[:strings.Index(image, ":")]

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

	buildArgs := data.GetString("build-args")

	if data.Has("dockerfile") {
		buildArgs += " --file " + data.GetString("dockerfile")
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
		tags = []string{image, fmt.Sprintf("%s:%v", prefix, "latest")}
		cacheFrom = "--cache-from=" + tags[1]
	} else {
		cacheFrom = "--cache-from=" + tags[0]
	}

	if err := utils.RunCmd(
		"docker build %s -t %s %s %s",
		buildArgs,
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

	if data.Has("images") && len(data.GetArray("images")) > 0 {
		for _, image := range data.GetArray("images") {
			if image.Has("branch") && image.GetString("branch") != data.GetStringOr("current-branch", "master") {
				if image.GetString("branch") != "*" {
					if m, _ := regexp.MatchString("^"+image.GetString("branch")+"$", data.GetStringOr("current-branch", "master")); !m {
						continue
					}
				}
			}

			workdir := ""
			if data.GetString("workdir") != "." {
				workdir = data.GetString("workdir") + "/"
			}

			if _, err := ioutil.ReadFile(workdir + image.GetString("dockerfile")); err != nil {
				continue
			}

			for k, v := range data.GetMap("/") {
				if k != "images" && !image.Has(k) {
					image.Set(k, v.Unwrap())
				}
			}

			fmt.Printf("\n")

			log.Printf("%s\n%s\n\n", color.GreenString(">>> build.docker-image sub-image:"), color.CyanString("%s", image.String()))

			if err := p.Run(image); err != nil {
				return err
			}

			fmt.Printf("\n")
		}
	}

	return nil
}
