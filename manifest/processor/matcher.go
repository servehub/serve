package processor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/servehub/utils/gabs"
	"github.com/servehub/utils/templater"
)

type Matcher struct{}

func (m Matcher) Process(tree *gabs.Container) error {
	return Repeat(5, func() error {
		return ProcessAll(tree, func(ktype string, output *gabs.Container, value interface{}, key interface{}) error {
			if ktype == "map" {
				skey, err := templater.Template(key.(string), tree)
				if err != nil {
					return fmt.Errorf("Error on template key `%v`: %v", key, err)
				}

				parts := strings.SplitN(skey, "?", 2)
				if valmap, ok := value.(map[string]interface{}); ok && len(parts) > 1 {
					matchValue := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
					newKey := strings.TrimSpace(parts[0])

					output.Delete(key.(string))
					output.Delete(newKey)

					if v, ok := valmap[matchValue]; ok {
						output.Set(v, newKey)
						return nil
					}

					for k, v := range valmap {
						if k == "*" {
							continue
						}

						re, err := regexp.Compile("^" + strings.Trim(k, "^$") + "$")
						if err != nil {
							return fmt.Errorf("Error on compile matcher regexp `%s`: %v", k, err)
						}

						matches := re.FindStringSubmatch(matchValue)
						groups := re.SubexpNames()

						if len(matches) > 0 {
							for i, group := range groups {
								if group != "" {
									tree.Set(matches[i], "match", group)
								}
							}

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
	})
}
