package templater

import (
	"testing"

	"github.com/InnovaCo/serve/utils/gabs"
)

type processorTestCase struct {
	in     string
	expect string
}

func TestUtilsTemplater(t *testing.T) {
	runAllProcessorTests(t, map[string]processorTestCase{
		"simple": {
			in:     `var`,
			expect: `var`,
		},

		"simple resolve with digit": {
			in:     `{{ var1 }}`,
			expect: `var1`,
		},

		"simple resolve with sep": {
			in:     `{{ var-var }}`,
			expect: `var-var`,
		},

		"simple resolve with dot": {
			in:     `{{ var.var }}`,
			expect: `1`,
		},

		"multi resolve": {
			in:     `{{ feature }}-{{ feature-suffix }}`,
			expect: `value-unknown-value-unknown`,
		},

		"replace": {
			in:     `{{ var--v |  replace('\W','_') }}`,
			expect: `var__v`,
		},

		"replace with whitespace": {
			in:     `{{ var--v | replace('\W',  '*') }}`,
			expect: `var**v`,
		},

		"multi resolve and replace": {
			in:     `{{ version | replace('\W',  '*') }}`,
			expect: `value*unknown*value*unknown`,
		},

		"multi resolve and replace with breaks": {
			in:     `{{ version | replace('[a-b]',  '*') }}`,
			expect: `v*lue-unknown-v*lue-unknown`,
		},

		"array value must print first element": {
			in:     `{{ list }}`,
			expect: `1`,
		},
	})
}

func runAllProcessorTests(t *testing.T, cases map[string]processorTestCase) {
	json := `{
		"var1": "var1",
		"var-var": "var-var",
		"var": {"var": "1"},
		"version": "{{ feature }}-{{ feature-suffix }}",
		"feature": "value-unknown",
		"feature-suffix": "{{ feature }}",
		"list": [1, 2, 3]
	}`

	tree, err := gabs.ParseJSON([]byte(json))
	if err != nil {
		t.Errorf("%v: failed!\n", err)
		t.Fail()
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			if res, err := Template(test.in, tree); err == nil {
				if test.expect != res {
					t.Errorf("%v: %v != %v: failed!\n", name, test.expect, res)
					t.Fail()
				}
			}
		})
	}
}
