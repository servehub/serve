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

func Template(s string, context *gabs.Container) (string, error) {
	t, err := fasttemplate.NewTemplate(s, "{{", "}}")
	if err != nil {
		return "", err
	}

	w := bytesBufferPool.Get().(*bytes.Buffer)

	if _, err := t.ExecuteFunc(w, func(w io.Writer, tag string) (int, error) {
		tag = strings.TrimSpace(tag)
		if value := context.Path(tag).Data(); value != nil {
			return w.Write([]byte(fmt.Sprintf("%v", value)))
		} else if v, err := ModifyExec(tag, context); err == nil {
			return w.Write([]byte(fmt.Sprintf("%v", v)))
		}else if strings.HasPrefix(tag, "vars.") || context.ExistsP(tag) {
			return 0, nil
		} else {
			return 0, fmt.Errorf("Undefined template variable: '%s'", tag)
		}
	}); err != nil {
		return "", err
	}

	out := string(w.Bytes())
	w.Reset()
	bytesBufferPool.Put(w)

	return out, nil
}