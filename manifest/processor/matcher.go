package processor

import (
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
)

type Matcher struct{}

func (m Matcher) Process(tree *gabs.Container) (*gabs.Container, error) {
	return ProcessAll(tree, func(ktype string, output *gabs.Container, value interface{}, key interface{}) error {
		if ktype == "map" {
			skey, err := template(key.(string), tree.Data())
			if err != nil {
				return err
			}

			parts := strings.SplitN(skey, "?", 2)
			if valmap, ok := value.(map[string]interface{}); ok && len(parts) > 1 {
				targetKey := strings.TrimSpace(parts[1])
				newKey := strings.TrimSpace(parts[0])

				output.Delete(key.(string))
				output.Set(nil, newKey)

				if v, ok := valmap[targetKey]; ok {
					output.Set(v, newKey)
					return nil
				}

				for k, v := range valmap {
					if ok, _ = regexp.MatchString("^" + strings.Trim(k, "^$") + "$", targetKey); ok {
						output.Set(v, newKey)
						return nil
					}
				}

				if v, ok := valmap["*"]; ok {
					output.Set(v, newKey)
				}
			}
		}

		return nil
	})
}
