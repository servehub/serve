package processor

import (
	"fmt"

	"github.com/Jeffail/gabs"
)

type Templater struct{}

func (t Templater) Process(tree *gabs.Container) (*gabs.Container, error) {
	var resp *gabs.Container
	var err error

	// magic: repeat N times for resolving all circular references
	for i := 1; i <= 3; i++ {
		resp, err = ProcessAll(tree, func (ktype string, output *gabs.Container, value interface{}, key interface{}) error {
			switch ktype {
			case "map":
				newKey, err := template(key.(string), tree.Data())
				if err != nil {
					return err
				}

				output.Delete(key.(string))
				output.Set(value, newKey)

			case "array":
				output.SetIndex(value, key.(int))

			default:
				newValue, err := template(fmt.Sprintf("%v", value), tree.Data())
				if err != nil {
					return err
				}

				output.Set(newValue)
			}

			return nil
		})
	}

	return resp, err
}
