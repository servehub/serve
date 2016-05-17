package manifest

import (
	"fmt"
	"io/ioutil"

	"github.com/Jeffail/gabs"
	"github.com/codegangsta/cli"
	"github.com/ghodss/yaml"
)

//manifest := LoadManifest(c)
//
//if manifest.Has("deploy") {
//	strategy, err := GetDeployStrategy(manifest.String("deploy.type", "default"))
//
//	if err != nil {
//		log.Fatalf("Unknown deploy strategy %s", strategy)
//	}
//
//	strategy.Release(manifest)
//}

//data, _ := ioutil.ReadFile("example/manifest.yml")
//jsonData, _ := yaml.YAMLToJSON(data)
//
//tree, _ := gabs.ParseJSON(jsonData)
//
//builds, _ := tree.Path("build").Children()
//for _, build := range builds {
//	if build.Exists("shell") {
//		println("shell ", build.Path("shell").Data().(string))
//	}
//
//	if build.Exists("marathon") {
//		println("marathon ", build.Path("marathon.package").Data().(string))
//	}
//
//	if build.Exists("debian") {
//		println("debian ", build.Path("debian").String())
//	}
//}

func LoadManifest(c *cli.Context) *Manifest {
	data, _ := ioutil.ReadFile("example/manifest.yml")
	jsonData, _ := yaml.YAMLToJSON(data)
	tree, _ := gabs.ParseJSON(jsonData)

	return &Manifest{tree: tree, ctx: c}
}

type Manifest struct {
	tree *gabs.Container
	ctx  *cli.Context
}

func (m Manifest) String() string {
	return m.tree.String()
}

func (m *Manifest) Has(path string) bool {
	return m.tree.ExistsP(path)
}

func (m *Manifest) Sub(path string) *Manifest {
	return &Manifest{m.tree.Path(path), m.ctx}
}

func (m *Manifest) Array(path string) []*Manifest {
	result := make([]*Manifest, 0)

	if chs, err := m.tree.Path(path).Children(); err == nil {
		for _, ch := range chs {
			result = append(result, &Manifest{ch, m.ctx})
		}
	}

	return result
}

func (m *Manifest) FirstKey() (string, error) {
	if res, err := m.tree.ChildrenMap(); err == nil {
		for k, _ := range res {
			return k, nil
		}
	}

	return "", fmt.Errorf("Object %v has no keys!", m)
}

func (m *Manifest) GetString(path string) string {
	return m.tree.Path(path).Data().(string)
}

func (m *Manifest) GetStringOr(path string, defaultVal string) string {
	if m.tree.ExistsP(path) {
		return m.tree.Path(path).Data().(string)
	} else {
		return defaultVal
	}
}

func (m *Manifest) ServiceName() string {
	return m.GetString("info.name")
}

func (m *Manifest) BuildVersion() string {
	return m.GetString("info.version")
}

