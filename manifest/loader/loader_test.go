package loader

import (
	"testing"
	"os"
	"io/ioutil"
	"fmt"

	"github.com/fatih/color"
)

func TestAnchorMerger(t *testing.T) {
	color.NoColor = false

	in := []byte(`---
vars: &v
  env: qa
deploy:
  env: live`,)
	expect := `{"deploy":{"env":"live"},"vars":{"env":"qa"}}`

    if err := ioutil.WriteFile("/tmp/test", in, 0644); err != nil {
		color.Red("Error file not create")
		t.Error("Error file not create")
		t.Fail()
	}

	defer os.Remove("/tmp/test")

	if g, err := LoadFile("/tmp/test"); err != nil {
		color.Red("%v\n", err)
		t.Error(err)
		t.Fail()
	} else {
		if g.String() == expect {
			color.Green("\n%s: OK\n", "FileLoad")
		} else {
			color.Red("%s != %s", g.String(), expect)
			t.Error(fmt.Errorf("%s != %s", g.String(), expect))
			t.Fail()
		}
	}
}
