package manifest

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"regexp"

	"github.com/InnovaCo/serve/utils/gabs"

	"github.com/InnovaCo/serve/manifest/loader"
	"github.com/InnovaCo/serve/manifest/processor"
)

type Manifest struct {
	tree *gabs.Container
}

func (m Manifest) String() string {
	return m.tree.StringIndent("", "  ")
}

func (m Manifest) Unwrap() interface{} {
	return m.tree.Data()
}

func (m Manifest) Has(path string) bool {
	v := m.tree.Path(path).Data()
	return v != nil && v != ""
}

func (m Manifest) GetString(path string) string {
	return fmt.Sprintf("%v", m.tree.Path(path).Data())
}

func (m Manifest) GetStringOr(path string, defaultVal string) string {
	if m.tree.ExistsP(path) {
		return m.GetString(path)
	} else {
		return defaultVal
	}
}

func (m Manifest) GetInt(path string) int {
	i, err := strconv.Atoi(m.GetString(path))
	if err != nil {
		log.Printf("Error on parse integer '%v' from: %v", path, m.GetString(path))
	}
	return i
}

func (m Manifest) GetBool(path string) bool {
	if v, ok := m.tree.Path(path).Data().(bool); ok {
		return v
	} else {
		return false
	}
}

func (m Manifest) GetMap(path string) map[string]Manifest {
	out := make(map[string]Manifest)
	mmap, err := m.tree.Path(path).ChildrenMap()
	if err != nil {
		log.Printf("Error get map '%v' from: %v", path, m.tree.Path(path).Data())
	}

	for k, v := range mmap {
		out[k] = Manifest{tree: v}
	}
	return out
}

func (m Manifest) GetArray(path string) []Manifest {
	out := make([]Manifest, 0)
	arr, err := m.tree.Path(path).Children()
	if err != nil {
		log.Printf("Error get array `%v` from: %v", path, m.tree.Path(path).Data())
	}

	for _, v := range arr {
		out = append(out, Manifest{tree: v})
	}
	return out
}

func (m Manifest) GetTree(path string) Manifest {
	return Manifest{tree: m.tree.Path(path)}
}

func (m Manifest) FindPlugins(plugin string) ([]PluginData, error) {
	tree := m.tree.Path(plugin)
	result := make([]PluginData, 0)

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

func (m Manifest) DelTree(path string) error {
	return m.tree.DeleteP(path)
}

func (m Manifest) GetPluginWithData(plugin string) PluginData {
	return makePluginPair(plugin, m.tree)
}

func (m Manifest) ToEnvArray(prefix string) map[string]string {
	result := make(map[string]string)
	if children, err := m.tree.ChildrenMap(); err == nil {
		for k, child := range children {
			reg, _ := regexp.Compile("\\W")
			c := Manifest{tree: child}.ToEnvArray(prefix + strings.ToUpper(string(reg.ReplaceAllString(k, "_"))) + "_")
			for k, v := range c {
				result[k] = v
			}
		}
	} else if children, err := m.tree.Children(); err == nil {
		for i, child := range children {
			c := Manifest{tree: child}.ToEnvArray(prefix + strconv.Itoa(i) + "_")
			for k, v := range c {
				result[k] = v
			}
		}
	} else {
		result[fmt.Sprintf("%s", prefix[:len(prefix) - 1])] = fmt.Sprintf("%v", m.tree.Data())
	}
	return result
}

func Load(path string, vars map[string]string) *Manifest {
	tree, err := loader.LoadFile(path)
	if err != nil {
		log.Fatalln("Error on load file:", err)
	}

	for k, v := range vars {
		tree.Set(v, "vars", k)
	}

	for _, proc := range processor.GetAll() {
		if err := proc.Process(tree); err != nil {
			log.Fatalf("Error in processor '%v': %v. \n\nManifest: %s", reflect.ValueOf(proc).Type().Name(), err, tree.StringIndent("", "  "))
		}
	}

	return &Manifest{tree: tree}
}

func LoadJSON(json string) *Manifest {
	tree, err := gabs.ParseJSON([]byte(json))
	if err != nil {
		log.Fatalf("Error on parse json '%s': %v\n", json, err)
	}

	return &Manifest{tree: tree}
}

func makePluginPair(plugin string, data *gabs.Container) PluginData {
	if s, ok := data.Data().(string); ok {
		obj := gabs.New()
		ns := strings.Split(plugin, ".")
		obj.Set(s, ns[len(ns)-1])
		data = obj
	}

	return PluginData{plugin, PluginRegestry.Get(plugin), Manifest{tree: data}}
}
