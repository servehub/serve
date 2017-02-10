package build

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.gradle", GradleBuild{})
}

type GradleBuild struct{}

func (p GradleBuild) Run(data manifest.Manifest) error {
	return utils.RunCmd(
		`docker run --rm -v "$PWD":/src -v ~/.gradle/caches/:/root/.gradle/caches/ -v ~/.gradle/wrapper/:/root/.gradle/wrapper/ -w /src frekele/gradle gradle %s -Pversion="%s"`,
		data.GetStringOr("gradle", ""),
		data.GetString("version"),
	)
}
