package manifest

import (
	"testing"
)

type testPlugin struct{}

func (t testPlugin) Run(_ Manifest) error {
	return nil
}

func TestPlugin(t *testing.T) {
	pd := pluginRegestry{}
	pd.Add("test_plugin", testPlugin{})
	exp := "test_plugin"

	t.Run("Has", func(t *testing.T) {
		if ok := pd.Has(exp); !ok {
			t.Errorf("Error: not has plugin %v", exp)
			t.Fail()
		}
	})

	t.Run("Get", func(t *testing.T) {
		if f := pd.Get("test_plugin"); f == nil {
			t.Errorf("Error: get plugin %v", exp)
			t.Fail()
		}
	})
}
