package processor

import (
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
)

func init() {
	ProcessorRegestry.Add("matcher", Matcher{})
}

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

				matched := false
				for k, v := range valmap {
					if k == targetKey || k == "*" {
						matched = true
					} else if strings.Contains(k, "*") {
						matched, err = regexp.MatchString(strings.Replace(k, "*", ".*", -1), targetKey)
						if err != nil {
							return err
						}
					}

					if matched {
						output.Set(v, newKey)
						break
					}
				}
			}
		}

		return nil
	})
}
