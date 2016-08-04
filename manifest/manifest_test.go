package manifest

import (
	"testing"

	"github.com/InnovaCo/serve/utils"
	"github.com/InnovaCo/serve/utils/gabs"
)

func TestManifest_ToEnvMap(t *testing.T) {
	tree, _ := gabs.ParseJSON([]byte(`
		{
			"info": { "name": "dima" },
			"foo": {
				"array": [
					10,
					20,
					30,
					"one",
					{ "sub": "obj" }
				]
			}
		}
	`))

	m := Manifest{tree}

	if !utils.MapsEqual(m.ToEnvMap("SERVE_"), map[string]string{
		"SERVE_FOO_ARRAY_0":     "10",
		"SERVE_FOO_ARRAY_1":     "20",
		"SERVE_INFO_NAME":       "dima",
		"SERVE_FOO_ARRAY_2":     "30",
		"SERVE_FOO_ARRAY_3":     "one",
		"SERVE_FOO_ARRAY_4_SUB": "obj",
	}) {
		t.Error("Error in ToEnvArray method")
	}
}
