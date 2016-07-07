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

	if jsonData, err := yaml.YAMLToJSON(data); err != nil {
		return nil, errors.New(color.RedString("Error on parse manifest: %v!", err))
	} else {
		return gabs.ParseJSON(jsonData)
	}
}
