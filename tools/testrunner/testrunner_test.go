package testrunner

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/fatih/color"
)

type testrunnerTestCase struct {
	first  map[string]interface{}
	second map[string]interface{}
	expect map[string]interface{}
}

func TestDiff(t *testing.T) {
	color.NoColor = false

	casses := map[string]testrunnerTestCase{
		"simple": {
			first:  map[string]interface{}{"name": "value"},
			second: map[string]interface{}{"name": "value"},
			expect: make(map[string]interface{}),
		},
		"diff": {
			first:  map[string]interface{}{"name1": "value"},
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
			color.Green("\nTest `%s`: OK\n", name)
		}
	}
}

func TestLoadData(t *testing.T) {
	color.NoColor = false

	in := []byte(`---
tests: plugin`)
	expect := map[string]interface{}{"tests": "plugin"}

	if err := ioutil.WriteFile("/tmp/test", in, 0644); err != nil {
		color.Red("Error file not create")
		t.Error("Error file not create")
		t.Fail()
	}

	defer os.Remove("/tmp/test")

	if data, err := loadData("/tmp/test"); err != nil {
		color.Red("%v\n", err)
		t.Error(err)
		t.Fail()
	} else {
		if d := diff(data, expect); !reflect.DeepEqual(d, make(map[string]interface{})) {
			color.Red("\n\nTest `load data` failed!")
			color.Yellow("\n\ndiff: %v\n", d)
			t.Fail()

		} else {
			color.Green("\nTest `load data`: OK\n")
		}
	}
}
