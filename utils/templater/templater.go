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
		p := strings.Split(tag, "|")
		path := strings.TrimSpace(p[0])
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