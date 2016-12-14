package plugins

import (
	"testing"
)

func TestExternalPlugin(t *testing.T) {
	runAllMultiCmdTests(t, map[string]processorTestCase{
		"simple": {
			in: `---
purge: false
ssh-user: "test_user"
target: "target_db_test"`,
			expect: map[string]interface{}{
				"cmdline": []string{""},
				},
		},
	},
	ExternalPlugin{})
}

