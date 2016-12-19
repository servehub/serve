package utils

import (
	"reflect"
	"testing"

	"github.com/fatih/color"
)

func TestUtils(t *testing.T) {
	color.NoColor = false

	t.Run("Substing", func(t *testing.T) {
		if s := Substr("test string", 2, 3); s != "st " {
			color.Red("Error in Substr")
			t.Errorf("Error %v != 'st '", s)
			t.Fail()
		}
	})

	t.Run("SubstingNotCorrectData", func(t *testing.T) {
		if s := Substr("test string", -2, 100); s != "test string" {
			color.Red("Error in Substr")
			t.Errorf("Error %v != 'test string'", s)
			t.Fail()
		}
	})

	em := map[string]string{"param1": "value1", "param2": "value2"}

	t.Run("MergeMaps", func(t *testing.T) {
		if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{"param2": "value2"}); !reflect.DeepEqual(m, em) {
			color.Red("Error in MergeMaps")
			t.Errorf("Error %v != %v", m, em)
			t.Fail()
		}
	})

	em = map[string]string{"param1": "value2"}

	t.Run("MergeMapsWithRewrite", func(t *testing.T) {
		if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{"param1": "value2"}); !reflect.DeepEqual(m, em) {
			color.Red("Error in MergeMaps")
			t.Errorf("Error %v != %v", m, em)
			t.Fail()
		}
	})

	em = map[string]string{"param1": "value1"}

	t.Run("MergeMapsWithEmpty", func(t *testing.T) {
		if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{}); !reflect.DeepEqual(m, em) {
			color.Red("Error in MergeMaps")
			t.Errorf("Error %v != %v", m, em)
			t.Fail()
		}
	})

	t.Run("MapsNotEqual", func(t *testing.T) {
		if MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value2"}) {
			color.Red("Error in MapsEqual")
			t.Fail()
		}
	})

	t.Run("MapsEqual", func(t *testing.T) {
		if !MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value1"}) {
			color.Red("Error in MapsEqual")
			t.Fail()
		}
	})

	t.Run("MapsNotEqualLen", func(t *testing.T) {
		if MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value1", "param2": "value2"}) {
			color.Red("Error in MapsEqual")
			t.Fail()
		}
	})

	t.Run("Contain", func(t *testing.T) {
		if !Contains("s", []string{"s", "t", "r"}) {
			color.Red("Error in Contains")
			t.Errorf("Error 's' != %v", []string{"s", "t", "r"})
			t.Fail()
		}
	})

	t.Run("ContainNot", func(t *testing.T) {
		if Contains("s", []string{"t", "r"}) {
			color.Red("Error in Contains")
			t.Errorf("Error 's' != %v", []string{"t", "r"})
			t.Fail()
		}
	})

	t.Run("RandomString", func(t *testing.T) {
		if l := RandomString(10); len(l) != 10 {
			color.Red("Error in RandomString")
			t.Errorf("Error %v != 10", len(l))
			t.Fail()
		}
	})
}
