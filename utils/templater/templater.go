package templater

import (
	"io"
	"strings"
	"fmt"
	"bytes"
	"sync"

	"github.com/valyala/fasttemplate"
	"github.com/InnovaCo/serve/utils/gabs"
)

var bytesBufferPool = sync.Pool{New: func() interface{} {
	return &bytes.Buffer{}
}}

const TIMES = 42

func Template(s string, context *gabs.Container) (string, error) {
	return _templater(s, context, true)
}

func MatchTemplate(s string, context *gabs.Container) (string, error) {
	return _templater(s, context, false)
}

func _templater(s string, context *gabs.Container, modify bool) (string, error) {
	var result = s
	var err error = nil

	for i := 0; i < TIMES; i++ {
		if result, err = _template(result, context, modify); err != nil {
			return "", err
		}
		if !(strings.Contains(result, "{{") && strings.Contains(result, "}}")) {
			break
		}
	}
	return result, nil
}

func _template(s string, context *gabs.Container, modify bool) (string, error) {
	t, err := fasttemplate.NewTemplate(s, "{{", "}}")
	if err != nil {
		return "", err
	}

	w := bytesBufferPool.Get().(*bytes.Buffer)

	if _, err := t.ExecuteFunc(w, func(w io.Writer, tag string) (int, error) {
		tag = strings.TrimSpace(tag)
		if value := context.Path(tag).Data(); value != nil {
			if valueArr, ok := value.([]interface{}); ok && len(valueArr) > 0 {
				value = valueArr[0]
			}

			return w.Write([]byte(fmt.Sprintf("%v", value)))
		} else if modify && strings.Contains(tag, "|") {
			if v, err := ModifyExec(tag, context); err == nil {
				return w.Write([]byte(fmt.Sprintf("%v", v)))
			}
		}

		if strings.HasPrefix(tag, "vars.") || context.ExistsP(tag) {
			return 0, nil
		}

		return 0, fmt.Errorf("Undefined template variable: '%s'", tag)
	}); err != nil {
		return "", err
	}

	out := string(w.Bytes())
	w.Reset()
	bytesBufferPool.Put(w)

	return out, nil
}
