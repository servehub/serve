package templater

import (
	"testing"

	"github.com/fatih/color"

	"github.com/InnovaCo/serve/utils/gabs"
	//"golang.org/x/crypto/ssh/test"
)

type processorTestCase struct {
	in     string
	expect string
}

func TestUtilsTemplater(t *testing.T) {
	runAllProcessorTests(t, map[string]processorTestCase{
		"simple": {
			in: `var`,
			expect: `var`,
		},

		"simple resolve with digit": {
			in: `{{ var1 }}`,
			expect: `var1`,
		},

		"simple resolve with sep": {
			in: `{{ var-var }}`,
			expect: `var-var`,
		},

		"simple resolve with dot": {
			in: `{{ var.var }}`,
			expect: `1`,
		},

		"multi resolve": {
			in: `{{ feature }}-{{ feature-suffix }}`,
			expect: `value-unknown-value-unknown`,
		},

		"replace": {
			in: `{{ var--v |  replace('\W','_') }}`,
			expect: `var__v`,
		},

		"replace with whitespace": {
			in: `{{ var--v | replace('\W',  '*') }}`,
			expect: `var**v`,
		},

		"multi resolve and replace": {
			in: `{{ version | replace('\W',  '*') }}`,
			expect: `value*unknown*value*unknown`,
		},
	})
}



func runAllProcessorTests(t *testing.T, cases map[string]processorTestCase) {
	color.NoColor = false

	json := `{
		"var1": "var1",
		"var-var": "var-var",
		"var": {"var": "1"},
		"version": "{{ feature }}-{{ feature-suffix }}",
		"feature": "value-unknown",
		"feature-suffix": "{{ feature }}"
	}`

	tree, err := gabs.ParseJSON([]byte(json))
	if err != nil {
		color.Red("%v: failed!\n", err)
		t.Fail()
	}

	for name, test := range cases {
		if res, err := Template(test.in, tree); err == nil {
			if test.expect == res {
				color.Green("%v: Ok\n", name)
			} else {
				color.Red("%v: %v != %v: failed!\n", name, test.expect, res)
				t.Fail()
			}
		} else {
			color.Red("error %v\n: failed!", err)
			t.Fail()
		}
	}
}
