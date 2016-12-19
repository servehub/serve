package loader

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestAnchorMerger(t *testing.T) {
	in := []byte(`---
vars: &v
  env: qa
deploy:
  env: live`)
	expect := `{"deploy":{"env":"live"},"vars":{"env":"qa"}}`

	if err := ioutil.WriteFile("/tmp/test", in, 0644); err != nil {
		t.Error("Error file not create")
		t.Fail()
	}

	defer os.Remove("/tmp/test")

	if g, err := LoadFile("/tmp/test"); err != nil {
		t.Errorf("%v\n", err)
		t.Fail()
	} else {
		if g.String() != expect {
			t.Errorf("%s != %s", g.String(), expect)
			t.Fail()
		}
	}
}
