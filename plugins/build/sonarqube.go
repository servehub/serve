package build

import (
	"io/ioutil"
	"os"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("build.sonarqube", SonarqubeBuild{})
}

type SonarqubeBuild struct{}

func (p SonarqubeBuild) Run(data manifest.Manifest) error {
	ioutil.WriteFile("sonar-project.properties", []byte(data.GetString("properties")), os.ModePerm)
	return manifest.PluginRegestry.Get("build.docker").Run(data)
}
