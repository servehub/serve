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

	t.Run("NotHas", func(t *testing.T) {
		exk := "count"
		if !m.Has(exk) {
			t.Errorf("Error: %v has %v", m.String(), exk)
			t.Fail()
		}
	})

	t.Run("Has", func(t *testing.T) {
		exk := "counter"
		if m.Has(exk) {
			t.Errorf("Error: %v not has %v", m.String(), exk)
			t.Fail()
		}
	})

	t.Run("ToEnvMap", func(t *testing.T) {
		exmp := map[string]string{
			"SERVE_FLAG":            "true",
			"SERVE_FOO_ARRAY_0":     "10",
			"SERVE_FOO_ARRAY_1":     "20",
			"SERVE_FOO_ARRAY_2":     "30",
			"SERVE_FOO_ARRAY_3":     "one",
			"SERVE_FOO_ARRAY_4_SUB": "obj",
			"SERVE_INFO_NAME":       "dima",
			"SERVE_COUNT":           "1",
		}
		if mp := m.ToEnvMap("SERVE_"); !utils.MapsEqual(mp, exmp) {
			t.Errorf("Error: %v != %v", mp, exmp)
			t.Fail()
		}
	})

	t.Run("GetInt", func(t *testing.T) {
		exi := 1
		if i := m.GetInt("count"); i != exi {
			t.Errorf("Error: %v != %v", i, exi)
			t.Fail()
		}
	})

	t.Run("GetIntOr", func(t *testing.T) {
		exi := 2
		if i := m.GetIntOr("counter", 2); i != exi {
			t.Errorf("Error: %v != %v", i, exi)
			t.Fail()
		}
	})

	t.Run("GetString", func(t *testing.T) {
		exs := "dima"
		if s := m.GetString("info.name"); s != exs {
			t.Errorf("Error: %v != %v", s, exs)
			t.Fail()
		}
	})

	t.Run("GetStringOr", func(t *testing.T) {
		exs := "dima"
		if s := m.GetStringOr("info.not_name", "dima"); s != exs {
			t.Errorf("Error: %v != %v", s, exs)
			t.Fail()
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		if b := m.GetBool("flag"); b != true {
			t.Errorf("Error: %v != %v", b, true)
			t.Fail()
		}
	})

	t.Run("GetTree", func(t *testing.T) {
		exs := "\"name\": \"dima\""
		if s := m.GetTree("info").String(); !strings.Contains(s, exs) {
			t.Errorf("Error: %v not contains %v", s, exs)
			t.Fail()
		}
	})

	t.Run("DelTree", func(t *testing.T) {
		if err := m.DelTree("info"); err != nil || strings.Contains(m.String(), "info") {
			t.Error("Error: not delete \"info\"")
			t.Fail()
		}
	})

}
