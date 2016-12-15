package plugins

import (
	"testing"
)

func TestExternalPlugin(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"simple": {
				in: `---
purge: false
param1: "value1"
param2:
  - "value1"
  - "value2"`,
				expect: map[string]interface{}{
					"cmdline": []string{"echo --plugin-data '{  \"param1\": \"value1\",  \"param2\": [    \"value1\",    \"value2\"  ],  \"purge\": false}'"},
				},
			},
			"empty": {
				in: `---`,
				expect: map[string]interface{}{
					"cmdline": []string{"echo --plugin-data '{}'"},
				},
			},
		},
		ExternalPlugin{"echo"})
}
