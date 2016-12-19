package manifest

import (
	"testing"

	"github.com/fatih/color"
)

type testPlugin struct{}

func (t testPlugin) Run(_ Manifest) error {
	return nil
}

func TestPlugin(t *testing.T) {
	color.NoColor = false

	pd := pluginRegestry{}
	pd.Add("test_plugin", testPlugin{})

	t.Run("Has", func(t *testing.T) {
		if ok := pd.Has("test_plugin"); !ok {
			color.Red("Error Has register plugin")
			t.Error("Error Has register plugin")
			t.Fail()
		}
	})

	t.Run("Get", func(t *testing.T) {
		if f := pd.Get("test_plugin"); f == nil {
			color.Red("Error Get register plugin")
			t.Error("Error Get register plugin")
			t.Fail()
		}
	})
}
