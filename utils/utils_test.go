package utils

import (
	"reflect"
	"testing"
)

func TestUtils(t *testing.T) {
	t.Run("Substing", func(t *testing.T) {
		if s := Substr("test string", 2, 3); s != "st " {
			t.Errorf("Error: %v != 'st '", s)
			t.Fail()
		}
	})

	t.Run("SubstingNotCorrectData", func(t *testing.T) {
		if s := Substr("test string", -2, 100); s != "test string" {
			t.Errorf("Error: %v != 'test string'", s)
			t.Fail()
		}
	})

	em := map[string]string{"param1": "value1", "param2": "value2"}

	t.Run("MergeMaps", func(t *testing.T) {
		if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{"param2": "value2"}); !reflect.DeepEqual(m, em) {
			t.Errorf("Error: %v != %v", m, em)
			t.Fail()
		}
	})

	em = map[string]string{"param1": "value2"}

	t.Run("MergeMapsWithRewrite", func(t *testing.T) {
		if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{"param1": "value2"}); !reflect.DeepEqual(m, em) {
			t.Errorf("Error: %v != %v", m, em)
			t.Fail()
		}
	})

	em = map[string]string{"param1": "value1"}

	t.Run("MergeMapsWithEmpty", func(t *testing.T) {
		if m := MergeMaps(map[string]string{"param1": "value1"}, map[string]string{}); !reflect.DeepEqual(m, em) {
			t.Errorf("Error: %v != %v", m, em)
			t.Fail()
		}
	})

	t.Run("MapsNotEqual", func(t *testing.T) {
		if MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value2"}) {
			t.Error("Error in MapsEqual")
			t.Fail()
		}
	})

	t.Run("MapsEqual", func(t *testing.T) {
		if !MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value1"}) {
			t.Error("Error in MapsEqual")
			t.Fail()
		}
	})

	t.Run("MapsNotEqualLen", func(t *testing.T) {
		if MapsEqual(map[string]string{"param1": "value1"}, map[string]string{"param1": "value1", "param2": "value2"}) {
			t.Error("Error in MapsEqual")
			t.Fail()
		}
	})

	t.Run("Contain", func(t *testing.T) {
		if !Contains("s", []string{"s", "t", "r"}) {
			t.Errorf("Error: 's' != %v", []string{"s", "t", "r"})
			t.Fail()
		}
	})

	t.Run("ContainNot", func(t *testing.T) {
		if Contains("s", []string{"t", "r"}) {
			t.Errorf("Error: 's' != %v", []string{"t", "r"})
			t.Fail()
		}
	})

	t.Run("RandomString", func(t *testing.T) {
		if l := RandomString(10); len(l) != 10 {
			t.Errorf("Error: %v != 10", len(l))
			t.Fail()
		}
	})
}
