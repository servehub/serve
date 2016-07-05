package manifest

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"

	"github.com/InnovaCo/serve/manifest/processor"
)

func Load(path string, vars map[string]string) *Manifest {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(color.RedString("Manifest file `%s` not found: %v", path, err))
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		log.Fatalln(color.RedString("Error on parse manifest: %v!", err))
	}

	tree, _ := gabs.ParseJSON(jsonData)

	for k, v := range vars {
		tree.Set(k, "vars", v)
	}

	for name, proc := range processor.ProcessorRegestry.GetAll() {
		tree, err = proc.Process(tree)
		if err != nil {
			log.Fatalf("Error in processor '%s': %v", name, err)
		}
	}

	return &Manifest{tree: tree}
}

type Manifest struct {
	tree *gabs.Container
}

func (m Manifest) String() string {
	return m.tree.String()
}

func (m Manifest) GetString(path string) string {
	return fmt.Sprintf("%v", m.tree.Path(path).Data())
}

func (m Manifest) FindPlugins(plugin string) ([]PluginPair, error) {
	tree := m.tree.Path(plugin)
	result := make([]PluginPair, 0)

	if tree.Data() == nil {
		return result, fmt.Errorf("Plugin '%s' not found in manifest", plugin)
	}

	if _, ok := tree.Data().([]interface{}); ok {
		arr, _ := tree.Children()
		for _, item := range arr {
			if _, ok := item.Data().(string); ok {
				result = append(result, makePluginPair(plugin, item))
			} else if res, err := item.ChildrenMap(); err == nil {
				for subplugin, data := range res {
					result = append(result, makePluginPair(plugin+"."+subplugin, data))
					break
				}
			}
		}
	} else {
		result = append(result, makePluginPair(plugin, tree))
	}

	return result, nil
}

func makePluginPair(plugin string, data *gabs.Container) PluginPair {
	if s, ok := data.Data().(string); ok {
		obj := gabs.New()
		ns := strings.Split(plugin, ".")
		obj.Set(s, ns[len(ns)-1])
		return PluginPair{plugin, PluginRegestry.Get(plugin), Manifest{obj}}
	} else {
		return PluginPair{plugin, PluginRegestry.Get(plugin), Manifest{data}}
	}
}
