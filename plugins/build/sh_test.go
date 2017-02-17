package build

import (
	"testing"

	"github.com/servehub/serve/tests"
)

func TestShBuild(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
					sh: "bash -c test.sh"
				`,
				Expects: []string{
					"bash -c test.sh",
				},
			},
		},
		ShBuild{})
}
