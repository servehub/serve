package manifest

import (
	"testing"
	"github.com/fatih/color"
)

type testPlugin struct{}

func (t testPlugin)Run(_ Manifest) error {
	return nil
}

func TestPlugin(t *testing.T) {
	color.NoColor = false

	pd := pluginRegestry{}
	pd.Add("test_plugin", testPlugin{})

	if ok := pd.Has("test_plugin"); ok {
		color.Green("\n%s: OK\n", "Has")
	} else {
		color.Red("Error Has register plugin")
		t.Error("Error Has register plugin")
		t.Fail()
	}

	if f := pd.Get("test_plugin"); f != nil  {
		color.Green("\n%s: OK\n", "Get")
	} else {
		color.Red("Error Get register plugin")
		t.Error("Error Get register plugin")
		t.Fail()
	}
}
