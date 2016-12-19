package manifest

import (
	"strings"
	"testing"

	"github.com/fatih/color"

	"github.com/InnovaCo/serve/utils"
	"github.com/InnovaCo/serve/utils/gabs"
)

func TestManifest(t *testing.T) {
	color.NoColor = false

	json := []byte(`
		{
			"info": { "name": "dima" },
			"count": 1,
			"flag": true,
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
	`)

	tree, _ := gabs.ParseJSON(json)

	m := Manifest{tree}

	t.Run("Has", func(t *testing.T) {
		if !m.Has("count") {
			color.Red("Error in Has method")
			t.Error("Error in Has method")
			t.Fail()
		}
	})

	t.Run("Has", func(t *testing.T) {
		if m.Has("counter") {
			color.Red("Error in Has method")
			t.Error("Error in Has method")
			t.Fail()
		}
	})

	t.Run("ToEnvMap", func(t *testing.T) {
		if !utils.MapsEqual(m.ToEnvMap("SERVE_"), map[string]string{
			"SERVE_FLAG":            "true",
			"SERVE_FOO_ARRAY_0":     "10",
			"SERVE_FOO_ARRAY_1":     "20",
			"SERVE_FOO_ARRAY_2":     "30",
			"SERVE_FOO_ARRAY_3":     "one",
			"SERVE_FOO_ARRAY_4_SUB": "obj",
			"SERVE_INFO_NAME":       "dima",
			"SERVE_COUNT":           "1",
		}) {
			color.Red("Error in ToEnvArray method")
			t.Error("Error in ToEnvArray method")
			t.Fail()
		}
	})

	t.Run("GetInt", func(t *testing.T) {
		if m.GetInt("count") != 1 {
			color.Red("Error in GetInt method")
			t.Error("Error in GetInt method")
			t.Fail()
		}
	})

	t.Run("GetIntOr", func(t *testing.T) {
		if m.GetIntOr("counter", 2) != 2 {
			color.Red("Error in GetIntOr method")
			t.Error("Error in GetIntOr method")
			t.Fail()
		}
	})

	t.Run("GetString", func(t *testing.T) {
		if m.GetString("info.name") != "dima" {
			color.Red("Error in GetString method")
			t.Error("Error in GetString method")
			t.Fail()
		}
	})

	t.Run("GetStringOr", func(t *testing.T) {
		if m.GetStringOr("info.not_name", "dima") != "dima" {
			color.Red("Error in GetStringOr method")
			t.Error("Error in GetStringOr method")
			t.Fail()
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		if m.GetBool("flag") != true {
			color.Red("Error in GetBool method")
			t.Error("Error in GetBool method")
			t.Fail()
		}
	})

	t.Run("GetTree", func(t *testing.T) {
		if !strings.Contains(m.GetTree("info").String(), "\"name\": \"dima\"") {
			color.Red("Error in GetTree method")
			t.Error("Error in GetTree method")
			t.Fail()
		}
	})

	t.Run("DelTree", func(t *testing.T) {
		if err := m.DelTree("info"); err != nil || strings.Contains(m.String(), "info") {
			color.Red("Error in DelTree method")
			t.Error("Error in DelTree method")
			t.Fail()
		}
	})

}
