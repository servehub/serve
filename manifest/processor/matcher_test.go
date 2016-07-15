package processor

import (
	"testing"

	"github.com/Jeffail/gabs"
)

func TestSimpleMatcher(t *testing.T) {
	jsonData := []byte(`
		{
			"vars": {
				"env": "qa"
			},
			"deploy": {
				"host ? {{ vars.env }}": {
					"qa": "qa-host.com",
					"live": "live-host.com"
				}
			}
		}
	`)

	tree, _ := gabs.ParseJSON(jsonData)

	proc := Matcher{}

	updated, err := proc.Process(tree)

	if err != nil {
		t.Fatal(err)
	}

	if updated.String() != `{"deploy":{"host":"qa-host.com"},"vars":{"env":"qa"}}` {
		t.Fatal("Unexpected result!", updated)
	}
}

func TestMatcherRegexpValue(t *testing.T) {
	jsonData := []byte(`
		{
			"vars": {
				"env": "qa-ru"
			},
			"deploy": {
				"host ? {{ vars.env }}": {
					"qa-.*": "qa-host.com",
					"live": "live-host.com"
				}
			}
		}
	`)

	tree, _ := gabs.ParseJSON(jsonData)

	proc := Matcher{}

	updated, err := proc.Process(tree)

	if err != nil {
		t.Fatal(err)
	}

	if updated.String() != `{"deploy":{"host":"qa-host.com"},"vars":{"env":"qa-ru"}}` {
		t.Fatal("Unexpected result!", updated)
	}
}

func TestMatcherWithDefaultValue(t *testing.T) {
	jsonData := []byte(`
		{
			"vars": {
				"env": "live-ru"
			},
			"deploy": {
				"host ? {{ vars.env }}": {
					"qa-.*": "qa-host.com",
					"live": "live-host.com",
					"*": "other"
				}
			}
		}
	`)

	tree, _ := gabs.ParseJSON(jsonData)

	proc := Matcher{}

	updated, err := proc.Process(tree)

	if err != nil {
		t.Fatal(err)
	}

	if updated.String() != `{"deploy":{"host":"other"},"vars":{"env":"live-ru"}}` {
		t.Fatal("Unexpected result!", updated)
	}
}

func TestMatcherReordering(t *testing.T) {
	jsonData := []byte(`
		{
			"vars": {
				"env": "live"
			},
			"deploy": {
				"host ? {{ vars.env }}": {
					"*": "other",
					"live": "live-host.com"
				}
			}
		}
	`)

	tree, _ := gabs.ParseJSON(jsonData)

	proc := Matcher{}

	updated, err := proc.Process(tree)

	if err != nil {
		t.Fatal(err)
	}

	if updated.String() != `{"deploy":{"host":"live-host.com"},"vars":{"env":"live"}}` {
		t.Fatal("Unexpected result!", updated)
	}
}

func TestMatcherReordering2(t *testing.T) {
	jsonData := []byte(`
		{
			"vars": {
				"env": "live"
			},
			"deploy": {
				"host ? {{ vars.env }}": {
					"live": "live-host.com",
					"*": "other"
				}
			}
		}
	`)

	tree, _ := gabs.ParseJSON(jsonData)

	proc := Matcher{}

	updated, err := proc.Process(tree)

	if err != nil {
		t.Fatal(err)
	}

	if updated.String() != `{"deploy":{"host":"live-host.com"},"vars":{"env":"live"}}` {
		t.Fatal("Unexpected result!", updated)
	}
}
