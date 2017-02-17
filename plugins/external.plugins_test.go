package plugins

import (
	"testing"
	"github.com/servehub/serve/tests"
)

func TestExternalPlugin(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"create": {
				In: `---
					purge: false
					param1: "value1"
					param2:
					  - "value1"
					  - "value2"
				`,
				Expects: []string{
					"echo --plugin-data '{  \"param1\": \"value1\",  \"param2\": [    \"value1\",    \"value2\"  ],  \"purge\": false}'",
				},
			},
			"empty": {
				In: `---`,
				Expects: []string{
					"echo --plugin-data '{}'",
				},
			},
		},
		ExternalPlugin{"echo"})
}
