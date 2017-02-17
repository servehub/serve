package build

import (
	"testing"

	"github.com/servehub/serve/tests"
)

func TestSbtBuild(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
					version: "0.0.0"
					test: "testOnly -- -l Integration"
				`,
				Expects: []string{
					"sbt ';set every version := \"0.0.0\"' clean \"testOnly -- -l Integration\" ",
				},
			},
		},
		SbtBuild{})
}
