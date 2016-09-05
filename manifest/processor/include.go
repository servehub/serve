package processor

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/InnovaCo/serve/manifest/loader"
	"github.com/InnovaCo/serve/utils/gabs"
	"github.com/InnovaCo/serve/utils/mergemap"
)

const ConfigPath = "/etc/serve"

type Include struct{}

func (in Include) Process(tree *gabs.Container) error {
	path := ConfigPath
	if path, ok := tree.Search("vars", "config-path").Data().(string); ok {
		path = path
	}

	if tree.ExistsP("include") {
		items, err := tree.Path("include").Children()
		if err != nil {
			return err
		}

		for _, inc := range items {
			if file, ok := inc.Search("file").Data().(string); ok {
				if !strings.HasPrefix(file, "/") {
					file = path + "/" + file
				}

				if err := includeFile(file, tree); err != nil {
					return err
				}
			}
		}
	}

	// include all base configs
	if files, err := filepath.Glob(path + "/conf.d/*.yml"); err == nil {
		for _, file := range files {
			if err := includeFile(file, tree); err != nil {
				return err
			}
		}
	}

	return nil
}

func includeFile(file string, tree *gabs.Container) error {
	loaded, err := loader.LoadFile(file)
	if err != nil {
		return err
	}

	log.Println("include:", file)

	merged, err := mergemap.Merge(loaded.Data().(map[string]interface{}), tree.Data().(map[string]interface{}))
	if err != nil {
		return err
	}

	_, err = tree.Set(merged)
	return err
}
