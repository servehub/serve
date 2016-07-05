package processor

import (
	"github.com/Jeffail/gabs"
	"github.com/InnovaCo/serve/manifest/loader"
	"github.com/InnovaCo/serve/utils/mergemap"
)

type Include struct{}

func (in Include) Process(tree *gabs.Container) (*gabs.Container, error) {
	return ProcessAll(tree, func(ktype string, output *gabs.Container, value interface{}, key interface{}) error {
		if ktype == "map" && key == "include" {
			items, err := output.Path("include").Children()
			if err != nil {
				return err
			}

			for _, inc := range items {
				if file, ok := inc.Search("file").Data().(string); ok {
					loaded, err := loader.LoadFile(file)
					if err != nil {
						return err
					}

					merged, err := mergemap.Merge(loaded.Data().(map[string]interface{}), output.Data().(map[string]interface{}))
					if err != nil {
						return err
					}

					output.Set(merged)
				}
			}
		}

		return nil
	})
}
