package processor

import (
	"github.com/Jeffail/gabs"
	"gopkg.in/flosch/pongo2.v3"
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
	Process(tree *gabs.Container) (*gabs.Container, error)
}

func ProcessAll(tree *gabs.Container, visitor func(string, *gabs.Container, interface{}, interface{}) error) (*gabs.Container, error) {
	if _, ok := tree.Data().(map[string]interface{}); ok {
		mmap, _ := tree.ChildrenMap()
		for k, v := range mmap {
			upd, err := ProcessAll(v, visitor)
			if err != nil {
				return nil, err
			}

			if err := visitor("map", tree, upd.Data(), k); err != nil {
				return nil, err
			}
		}
	} else if _, ok := tree.Data().([]interface{}); ok {
		marray, _ := tree.Children()
		for i, v := range marray {
			upd, err := ProcessAll(v, visitor)
			if err != nil {
				return nil, err
			}

			if err := visitor("array", tree, upd.Data(), i); err != nil {
				return nil, err
			}
		}
	} else {
		if err := visitor("other", tree, tree.Data(), nil); err != nil {
			return nil, err
		}
	}

	return tree, nil
}

func template(s string, context interface{}) (string, error) {
	tpl, err := pongo2.FromString(s)
	if err != nil {
		return "", err
	}

	ctx := pongo2.Context{}
	if m, ok := context.(map[string]interface{}); ok {
		ctx = m
	}

	out, err := tpl.Execute(ctx)

	if err != nil {
		return "", err
	}

	return out, nil
}
