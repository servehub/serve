package processor

import (
	"fmt"

	"github.com/InnovaCo/serve/utils/gabs"
)

type Templater struct{}

func (t Templater) Process(tree *gabs.Container) error {
	return Repeat(3, func() error {
		return ProcessAll(tree, func (ktype string, output *gabs.Container, value interface{}, key interface{}) error {
			switch ktype {
			case "map":
				newKey, err := template(key.(string), tree)
				if err != nil {
					return err
				}

				output.Delete(key.(string))
				output.Set(value, newKey)

			case "array":
				output.SetIndex(value, key.(int))

			default:
				switch value.(type) {
				case bool:
				case int:
				case int32:
				case int64:
				case float32:
				case float64:
				case nil:
					output.Set(value)
				default:
					newValue, err := template(fmt.Sprintf("%v", value), tree)
					if err != nil {
						return err
					}

					output.Set(newValue)
				}
			}

			return nil
		})
	})
}
