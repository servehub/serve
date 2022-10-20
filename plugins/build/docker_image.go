package build

import (
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
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
	workdir := data.GetString("workdir") + "/"

	if data.Has("dockerfile") {
		buildArgs += " --file " + workdir + data.GetString("dockerfile")
	}

	tags := make([]string, 0)
	for _, tag := range data.GetArrayForce("tags") {
		tags = append(tags, fmt.Sprintf("%s:%v", prefix, tag))
	}

	labels := make([]string, 0)
	for labelName, labelValue := range data.GetMap("labels") {
		if labelValue.Unwrap() != "" {
			labels = append(labels, fmt.Sprintf(` --label "%s=%s"`, labelName, labelValue.Unwrap()))
		}
	}

	if len(labels) > 0 {
		sort.Strings(labels)
		buildArgs += strings.Join(labels, "")
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

	envs := make(map[string]string)
	for k, v := range data.GetMap("environment") {
		envs[k] = strings.TrimSpace(fmt.Sprintf("%v", v.Unwrap()))
	}

	if err := utils.RunCmdWithEnv(
		fmt.Sprintf(
			"docker build %s -t %s %s %s",
			buildArgs,
			strings.Join(tags, " -t "),
			cacheFrom,
			data.GetString("workdir"),
		),
		envs,
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
				if image.GetBool("skip-errors") {
					log.Printf("Error on build sub-image: %v", err)
				} else {
					return err
				}
			}

			fmt.Printf("\n")
		}
	}

	return nil
}
