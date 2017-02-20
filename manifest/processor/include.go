package processor

import (
	"path/filepath"
	"strings"

	"github.com/servehub/serve/manifest/config"
	"github.com/servehub/utils/gabs"
)

const ConfigPath = "/etc/serve"

type Include struct{}

func (in Include) Process(tree *gabs.Container) error {
	path := ConfigPath
	if customPath, ok := tree.Path("vars.config-path").Data().(string); ok {
		path = customPath
	}

	if tree.ExistsP("include") {
		items, err := tree.Path("include").Children()
		if err != nil {
			return err
		}

		for i, _ := range items {
			inc := items[len(items)-i-1] // loop in reverse order for merge priority

			if file, ok := inc.Search("file").Data().(string); ok {
				if !strings.HasPrefix(file, "/") {
					file = path + "/" + file
				}

				if err := tree.WithFallbackYamlFile(file); err != nil {
					return err
				}
			}
		}
	}

	if files, err := filepath.Glob(path + "/conf.d/*.yml"); err == nil {
		for _, file := range files {
			if err := tree.WithFallbackYamlFile(file); err != nil {
				return err
			}
		}
	}

	return tree.WithFallbackYaml(config.MustAsset("config/reference.yml"))
}
