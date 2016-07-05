package loader

import (
	"errors"
	"io/ioutil"

	"github.com/Jeffail/gabs"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
)

func LoadFile(path string) (*gabs.Container, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New(color.RedString("Manifest file `%s` not found: %v", path, err))
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, errors.New(color.RedString("Error on parse manifest: %v!", err))
	}

	tree, _ := gabs.ParseJSON(jsonData)

	return tree, nil
}
