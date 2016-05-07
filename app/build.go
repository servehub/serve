package app

import (
	"github.com/codegangsta/cli"
	"github.com/ghodss/yaml"
	"github.com/Jeffail/gabs"
	"io/ioutil"
)

func BuildCommand() cli.Command {
	return cli.Command{
		Name:  "build",
		Usage: "Duild package",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "branch"},
			cli.StringFlag{Name: "build-number"},
		},
		Action: func(c *cli.Context) {

			data, _ := ioutil.ReadFile("example/manifest.yml")
			jsonData, _ := yaml.YAMLToJSON(data)

			tree, _ := gabs.ParseJSON(jsonData)

			builds, _ := tree.Path("build").Children()
			for _, build := range builds {
				if build.Exists("shell") {
					println("shell ", build.Path("shell").Data().(string))
				}

				if build.Exists("marathon") {
					println("marathon ", build.Path("marathon.package").Data().(string))
				}

				if build.Exists("debian") {
					println("debian ", build.Path("debian").String())
				}
			}
		},
	}
}

//type Manifest struct {
//	conf *viper.Viper
//}
//
//func (m *Manifest) GetString(key string) string {
//	return m.conf.GetString(key)
//}
