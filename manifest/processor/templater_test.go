package processor

import (
	"testing"
	"github.com/Jeffail/gabs"
)

func TestTemplater(t *testing.T) {
	jsonData := []byte(`
		{
			"info": {
				"name": "{{ vars.hello }}-name",
				"descr ? {{ vars.env }}": "my supa proj"
			},
			"vars": {
				"hello": "Hello, World!11",
				"env": "qa"
			},
			"builds": [
				{ "name": "Kulikov", "full-name": "{{ builds.0.name }} Dmitry" }
			],
			"test": {
				"super_peper_key": "hi!",
				"non-exists-var": "{{ test.super_peper_key }}"
			}
		}
	`)

	tree, _ := gabs.ParseJSON(jsonData)

	proc := Templater{}

	updated, err := proc.Process(tree)

	if err != nil {
		t.Fatal(err)
	}

	if updated.String() != `{"builds":[{"full-name":"Kulikov Dmitry","name":"Kulikov"}],"info":{"descr ? qa":"my supa proj","name":"Hello, World!11-name"},"test":{"non-exists-var":"hi!","super_peper_key":"hi!"},"vars":{"env":"qa","hello":"Hello, World!11"}}` {
		t.Log(updated)
		t.Fatal("Unexpected result!")
	}
}

func TestNonScalarVars(t *testing.T) {
	jsonData := []byte(`
		{
			"items": ["1", "2", "3"],
			"output": "{{ items.0 }}"
		}
	`)

	tree, _ := gabs.ParseJSON(jsonData)

	proc := Templater{}

	updated, err := proc.Process(tree)

	if err != nil {
		t.Fatal(err)
	}

	if updated.String() != `{"items":["1","2","3"],"output":"1"}` {
		t.Log(updated)
		t.Fatal("Unexpected result!")
	}
}
