package manifest

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/servehub/serve/manifest/processor"
	"github.com/servehub/utils"
	"github.com/servehub/utils/gabs"
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
	}
	return defaultVal
}

func (m Manifest) GetFloat(path string) float64 {
	f, err := strconv.ParseFloat(m.GetString(path), 64)
	if err != nil {
		log.Fatalf("Error on parse float64 '%v' from: %v", path, m.GetString(path))
	}
	return f
}

func (m Manifest) GetInt(path string) int {
	i, err := strconv.Atoi(m.GetString(path))
	if err != nil {
		log.Fatalf("Error on parse integer '%v' from: %v", path, m.GetString(path))
	}
	return i
}

func (m Manifest) GetIntOr(path string, defaultVal int) int {
	if m.tree.ExistsP(path) {
		return m.GetInt(path)
	}
	return defaultVal
}

func (m Manifest) GetBool(path string) bool {
	return strings.ToLower(m.GetString(path)) == "true"
}

func (m Manifest) GetMap(path string) map[string]Manifest {
	out := make(map[string]Manifest)

	tree := m.tree
	if len(path) > 0 && path != "." && path != "/" {
		tree = m.tree.Path(path)
	}

	mmap, err := tree.ChildrenMap()
	if err != nil {
		log.Fatalf("Error get map '%v' from: %v. Error: %s", path, m.tree.Path(path).Data(), err)
	}

	for k, v := range mmap {
		out[k] = Manifest{v}
	}
	return out
}

func (m Manifest) GetArray(path string) []Manifest {
	out := make([]Manifest, 0)
	arr, err := m.tree.Path(path).Children()
	if err != nil {
		log.Fatalf("Error get array `%v` from: %v", path, m.tree.Path(path).Data())
	}

	for _, v := range arr {
		out = append(out, Manifest{v})
	}
	return out
}

func (m Manifest) GetArrayForce(path string) []interface{} {
	out := make([]interface{}, 0)

	arr, err := m.tree.Path(path).Children()
	if err != nil && m.tree.ExistsP(path) {
		arr = append(arr, m.tree.Path(path))
	}

	for _, v := range arr {
		out = append(out, v.Data())
	}

	return out
}

func (m Manifest) GetTree(path string) Manifest {
	return Manifest{m.tree.Path(path)}
}

func (m Manifest) Set(path string, value interface{}) {
	m.tree.SetP(value, path)
}

func (m Manifest) ArrayAppend(path string, value interface{}) {
	m.tree.ArrayAppendP(value, path)
}

func (m Manifest) FindPlugins(plugin string) ([]PluginData, error) {
	tree := m.tree.Path(plugin)
	result := make([]PluginData, 0)

	if tree.Data() == nil {
		return nil, fmt.Errorf("Plugin `%s` not found in manifest!", plugin)
	}

	if _, ok := tree.Data().([]interface{}); ok {
		arr, _ := tree.Children()
		for _, item := range arr {
			if _, ok := item.Data().(string); ok {
				result = append(result, makePluginPair(plugin, item))
			} else if res, err := item.ChildrenMap(); err == nil {
				if len(res) == 1 {
					for subplugin, data := range res {
						result = append(result, makePluginPair(plugin+"."+subplugin, data))
						break
					}
				} else if len(res) == 0 && !PluginRegestry.Has(plugin) {
					// skip subplugin with empty data
				} else {
					result = append(result, makePluginPair(plugin, item))
				}
			}
		}
	} else if PluginRegestry.Has(plugin) {
		result = append(result, makePluginPair(plugin, tree))
	} else {
		log.Println(color.YellowString("Plugins for `%s` section not specified, skip...", plugin))
	}

	return result, nil
}

func (m Manifest) DelTree(path string) error {
	return m.tree.DeleteP(path)
}

func (m Manifest) GetPluginWithData(plugin string) PluginData {
	return makePluginPair(plugin, m.tree)
}

var envNameRegex = regexp.MustCompile("\\W")

func (m Manifest) ToEnvMap(prefix string) map[string]string {
	result := make(map[string]string)
	if children, err := m.tree.ChildrenMap(); err == nil {
		for k, child := range children {
			result = utils.MergeMaps(result, Manifest{child}.ToEnvMap(prefix+strings.ToUpper(envNameRegex.ReplaceAllString(k, "_"))+"_"))
		}
	} else if children, err := m.tree.Children(); err == nil {
		for i, child := range children {
			result = utils.MergeMaps(result, Manifest{child}.ToEnvMap(prefix+strconv.Itoa(i)+"_"))
		}
	} else if m.tree.Data() != nil {
		result[prefix[:len(prefix)-1]] = fmt.Sprintf("%v", m.tree.Data())
	}
	return result
}

func Load(path string, vars map[string]string) *Manifest {
	tree, err := gabs.LoadYamlFile(path)
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

	return &Manifest{tree}
}

func LoadJSON(path string) *Manifest {
	tree, err := gabs.ParseJSONFile(path)
	if err != nil {
		log.Fatalf("Error on load json file '%s': %v\n", path, err)
	}

	return &Manifest{tree}
}

func ParseJSON(json string) *Manifest {
	tree, err := gabs.ParseJSON([]byte(json))
	if err != nil {
		log.Fatalf("Error on parse json '%s': %v\n", json, err)
	}

	return &Manifest{tree}
}

func makePluginPair(plugin string, data *gabs.Container) PluginData {
	if s, ok := data.Data().(string); ok {
		obj := gabs.New()
		ns := strings.Split(plugin, ".")
		obj.Set(s, ns[len(ns)-1])
		data = obj
	} else {
		var cpy interface{}
		bs, _ := json.Marshal(data.Data())
		json.Unmarshal(bs, &cpy)
		data.Set(cpy)
	}

	return PluginData{plugin, PluginRegestry.Get(plugin), Manifest{data}}
}
