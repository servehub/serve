package processor

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/servehub/serve/manifest/config"
	"github.com/servehub/utils/gabs"
)

const DefaultConfigPath = "/etc/serve"

type Include struct{}

func (in Include) Process(tree *gabs.Container) error {
	path := DefaultConfigPath

	if envPath, ok := os.LookupEnv("SERVE_CONFIG_PATH"); ok {
		path = envPath
	}

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
				} else if path != DefaultConfigPath && strings.HasPrefix(file, DefaultConfigPath) {
					file = strings.Replace(file, DefaultConfigPath, path, 1) // if file has absolute path to default config-dir — replace to custom
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
