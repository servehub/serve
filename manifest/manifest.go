package manifest

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Jeffail/gabs"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
)

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

func (m Manifest) flag(name string) string {
	if res := m.ctx.String(name); res != "" {
		return res
	} else {
		return m.ctx.GlobalString(name)
	}
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
	if m.tree.ExistsP(path) {
		d := m.tree.Path(path).Data()

		if obj, ok := d.(map[string]interface{}); ok {
			if v, ok := obj[m.flag("env")]; ok {
				d = v
			} else {
				log.Fatalln(color.RedString("manifest: not found '%s' in %s", m.flag("env"), m.tree.Path(path).String()))
			}
		}

		return fmt.Sprintf("%v", d)
	} else {
		log.Fatalln(color.RedString("manifest: path `%s` not found in %v", path, m))
		return ""
	}
}

func (m *Manifest) GetStringOr(path string, defaultVal string) string {
	if m.tree.ExistsP(path) {
		return m.GetString(path)
	} else {
		return defaultVal
	}
}

func (m *Manifest) ServiceName() string {
	return m.GetString("info.name")
}

func (m *Manifest) BuildVersion() string {
	return fmt.Sprintf("%s.%s", m.GetString("info.version"), m.flag("build-number"))
}
