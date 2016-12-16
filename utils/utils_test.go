package utils

import (
	"reflect"
	"testing"

	"github.com/fatih/color"
)

func TestUtilits(t *testing.T) {
	color.NoColor = false

	if s := Substr("test string", 2, 3); s != "st " {
		color.Red("Error in Substr")
		t.Errorf("Error %v != 'st '", s)
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "substing")
	}

	if s := Substr("test string", -2, 100); s != "test string" {
		color.Red("Error in Substr")
		t.Errorf("Error %v != 'test string'", s)
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "substing non correct data")
	}

	em := map[string]string{"param1": "value1", "param2": "value2"}

	if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{"param2": "value2"}); !reflect.DeepEqual(m, em) {
		color.Red("Error in MergeMaps")
		t.Errorf("Error %v != %v", m, em)
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "merge maps")
	}

	em = map[string]string{"param1": "value2"}

	if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{"param1": "value2"}); !reflect.DeepEqual(m, em) {
		color.Red("Error in MergeMaps")
		t.Errorf("Error %v != %v", m, em)
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "merge maps with rewrite")
	}

	em = map[string]string{"param1": "value1"}

	if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{}); !reflect.DeepEqual(m, em) {
		color.Red("Error in MergeMaps")
		t.Errorf("Error %v != %v", m, em)
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "merge maps with empty")
	}

	if MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value2"}) {
		color.Red("Error in MapsEqual")
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "maps not equal")
	}

	if !MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value1"}) {
		color.Red("Error in MapsEqual")
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "maps equal")
	}

	if MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value1", "param2": "value2"}) {
		color.Red("Error in MapsEqual")
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "maps not equal len")
	}

	if !Contains("s", []string{"s", "t", "r"}) {
		color.Red("Error in Contains")
		t.Errorf("Error 's' != %v", []string{"s", "t", "r"})
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "contain")
	}

	if Contains("s", []string{"t", "r"}) {
		color.Red("Error in Contains")
		t.Errorf("Error 's' != %v", []string{"t", "r"})
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "not contain")
	}

	if l := RandomString(10); len(l) != 10 {
		color.Red("Error in RandomString")
		t.Errorf("Error %v != 10", len(l))
		t.Fail()
	} else {
		color.Green("\n%s: OK\n", "random string")
	}
}
