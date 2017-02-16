package loader

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"

	"github.com/servehub/serve/utils/gabs"
)

func LoadFile(path string) (*gabs.Container, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New(color.RedString("Manifest file `%s` not found: %v", path, err))
	}

	if result, err := ParseYaml(data); err != nil {
		return nil, fmt.Errorf("Error on load file %s: %v", path, err)
	} else {
		return result, nil
	}
}

func ParseYaml(data []byte) (*gabs.Container, error) {
	if jsonData, err := yaml.YAMLToJSON(data); err != nil {
		return nil, errors.New(color.RedString("Error on parse yaml: %v", err))
	} else {
		return gabs.ParseJSON(jsonData)
	}
}
