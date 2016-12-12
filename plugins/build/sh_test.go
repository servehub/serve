package build

import (
	"testing"
)

func TestShBuild(t *testing.T) {
	runAllMultiCmdTests(t, map[string]processorTestCase{
		"simple": {
			in: `---
sh: "bash -c test.sh"`,
			expect: map[string]interface{}{
				"cmdline": []string{"bash -c test.sh"},
				},
		},
	},
	ShBuild{})
}