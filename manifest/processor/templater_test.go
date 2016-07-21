package processor

import (
	"testing"

	"github.com/fatih/color"

	"github.com/InnovaCo/serve/utils/gabs"
)

func TestTemplater(t *testing.T) {
	runAllProcessorTests(t, func() Processor { return Templater{} }, map[string]processorTestCase{
		"simple template": {
			in: `
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
			`,
			expect: `{"builds":[{"full-name":"Kulikov Dmitry","name":"Kulikov"}],"info":{"descr ? qa":"my supa proj","name":"Hello, World!11-name"},"test":{"non-exists-var":"hi!","super_peper_key":"hi!"},"vars":{"env":"qa","hello":"Hello, World!11"}}`,
		},

		"non scalar vars": {
			in: `
				{
					"items": ["1", "2", "3"],
					"output": "{{ items.0 }}"
				}
			`,
			expect: `{"items":["1","2","3"],"output":"1"}`,
		},
	})
}

func runAllProcessorTests(t *testing.T, processor func() Processor, cases map[string]processorTestCase) {
	color.NoColor = false

	for name, test := range cases {
		tree, _ := gabs.ParseJSON([]byte(test.in))
		proc := processor()
		err := proc.Process(tree)

		if err != nil {
			t.Fatal(err)
		}

		if tree.String() != test.expect {
			color.Red("\n\nTest `%s` failed!", name)
			color.Yellow("\n\nexpected:  %s\n\nbut given: %s\n\n", test.expect, tree.String())
			t.Fail()
		} else {
			color.Green("\n%s: OK\n", name)
		}
	}
}

type processorTestCase struct {
	in     string
	expect string
}
