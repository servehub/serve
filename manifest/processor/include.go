package processor

import (
	"path/filepath"
	"strings"
	"log"

	"github.com/InnovaCo/serve/utils/gabs"
	"github.com/InnovaCo/serve/manifest/loader"
	"github.com/InnovaCo/serve/utils/mergemap"
)

const (
	ConfdPath    string = "/etc/serve/conf.d"
	IncludedPath string = "/etc/serve/include.d"
)

type Include struct{}

func (in Include) Process(tree *gabs.Container) error {
	if tree.ExistsP("include") {
		items, err := tree.Path("include").Children()
		if err != nil {
			return err
		}

		for _, inc := range items {
			if file, ok := inc.Search("file").Data().(string); ok {
				if !strings.HasPrefix(file, "/") {
					file = IncludedPath + "/" + file
				}

				if err := includeFile(file, tree); err != nil {
					return err
				}
			}
		}
	}

	// include all base configs
	if files, err := filepath.Glob(ConfdPath + "/*.yml"); err == nil {
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

	log.Println("include: ", file)

	merged, err := mergemap.Merge(loaded.Data().(map[string]interface{}), tree.Data().(map[string]interface{}))
	if err != nil {
		return err
	}

	_, err = tree.Set(merged)
	return err
}
