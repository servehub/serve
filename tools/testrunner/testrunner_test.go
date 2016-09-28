package testrunner

import (
	"testing"

	"github.com/fatih/color"
	"reflect"
)


type testrunnerTestCase struct {
	first   map[string]interface{}
	second  map[string]interface{}
	expect  map[string]interface{}
}

func TestPlugin(t *testing.T) {
	color.NoColor = false

	casses := map[string]testrunnerTestCase{
		"simple":
			{
				first: map[string]interface{}{"name": "value"},
				second: map[string]interface{}{"name": "value"},
				expect:  make(map[string]interface{}),
			},
		"diff":
			{
				first: map[string]interface{}{"name1": "value"},
				second: map[string]interface{}{"name": "value"},
				expect: map[string]interface{}{"name": "<nil> != value", "name1": "value != <nil>"},
			},
	}

	for name, test := range casses {
		if d := diff(test.first, test.second); !reflect.DeepEqual(d, test.expect) {
			color.Red("\n\nTest `%s` failed!", name)
			color.Yellow("\n\nexpected:  %v\n\ngiven: %v\n\n", test.expect, d)
			t.Fail()
		} else {
			color.Green("\n%s: OK\n", name)
		}
	}

}