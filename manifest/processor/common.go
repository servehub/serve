package processor

import (
	"fmt"
	"strings"

	"github.com/InnovaCo/serve/utils/gabs"
)

func GetAll() []Processor {
	return []Processor{
		Include{},
		Matcher{},
		AnchorMerger{},
		Templater{},
	}
}

type Processor interface {
	Process(tree *gabs.Container) error
}

func ProcessAll(tree *gabs.Container, visitor func(string, *gabs.Container, interface{}, interface{}) error) error {
	errors := make([]string, 0)

	if _, ok := tree.Data().(map[string]interface{}); ok {
		mmap, _ := tree.ChildrenMap()
		for k, v := range mmap {
			if err := ProcessAll(v, visitor); err != nil {
				errors = append(errors, err.Error())
				continue
			}

			if err := visitor("map", tree, v.Data(), k); err != nil {
				errors = append(errors, err.Error())
			}
		}
	} else if _, ok := tree.Data().([]interface{}); ok {
		marray, _ := tree.Children()
		for i, v := range marray {
			if err := ProcessAll(v, visitor); err != nil {
				errors = append(errors, err.Error())
				continue
			}

			if err := visitor("array", tree, v.Data(), i); err != nil {
				errors = append(errors, err.Error())
			}
		}
	} else {
		if err := visitor("other", tree, tree.Data(), nil); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if (len(errors) == 0) {
		return nil
	} else {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}
}

/**
 * magic: repeat N times for resolving all circular references
 */
func Repeat(n int, f func() error) error {
	var err error
	for i := 0; i < n; i++ {
		err = f()
	}
	return err
}
