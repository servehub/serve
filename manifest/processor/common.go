package processor

import (
	"bytes"
	"io"
	"sync"
	"fmt"
	"strings"

	"github.com/valyala/fasttemplate"

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

var bytesBufferPool = sync.Pool{New: func() interface{} {
	return &bytes.Buffer{}
}}

func template(s string, context *gabs.Container) (string, error) {
	t, err := fasttemplate.NewTemplate(s, "{{", "}}")
	if err != nil {
		return "", err
	}

	w := bytesBufferPool.Get().(*bytes.Buffer)

	if _, err := t.ExecuteFunc(w, func(w io.Writer, tag string) (int, error) {
		path := strings.TrimSpace(tag)
		if value := context.Path(path).Data(); value != nil {
			return w.Write([]byte(fmt.Sprintf("%v", value)))
		} else if strings.HasPrefix(path, "vars.") || context.ExistsP(path) {
			return 0, nil
		} else {
			return 0, fmt.Errorf("Undefined template variable: '%s'", path)
		}
	}); err != nil {
		return "", err
	}

	out := string(w.Bytes())
	w.Reset()
	bytesBufferPool.Put(w)

	return out, nil
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
