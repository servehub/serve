package build

import (
	"testing"
)

func TestSbtBuild(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"simple": {
				in: `---
version: "0.0.0"
test: "testOnly -- -l Integration"`,
				expect: map[string]interface{}{
					"cmdline": []string{"sbt ';set every version := \"0.0.0\"' clean \"testOnly -- -l Integration\" "},
				},
			},
		},
		SbtBuild{})
}
