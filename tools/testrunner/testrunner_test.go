package testrunner

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type testrunnerTestCase struct {
	first  map[string]interface{}
	second map[string]interface{}
	expect map[string]interface{}
}

func TestDiff(t *testing.T) {
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
		t.Run(name, func(t *testing.T) {
			if d := diff(test.first, test.second); !reflect.DeepEqual(d, test.expect) {
				t.Errorf("Error:\nexpected:  %v\n\ngiven: %v\n\n", test.expect, d)
				t.Fail()
			}
		})
	}
}

func TestLoadData(t *testing.T) {
	in := []byte(`---
tests: plugin`)
	expect := map[string]interface{}{"tests": "plugin"}

	if err := ioutil.WriteFile("/tmp/test", in, 0644); err != nil {
		t.Error("Error file not create")
		t.Fail()
	}

	defer os.Remove("/tmp/test")

	if data, err := loadData("/tmp/test"); err != nil {
		t.Errorf("%v\n", err)
		t.Fail()
	} else {
		if d := diff(data, expect); !reflect.DeepEqual(d, make(map[string]interface{})) {
			t.Errorf("\n\ndiff: %v\n", d)
			t.Fail()
		}
	}
}
