package build

import (
	"fmt"
	"testing"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func TestMarathonBuild(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"simple": {
				in: `---
source: target/pack
registry-url: test.ru`,
				expect: map[string]interface{}{
					"cmdline": []string{"tar -zcf marathon.tar.gz -C target/pack/ .",
						"curl -vsSf -XPUT -T marathon.tar.gz test.ru",
					},
				},
			},
		},
		MarathonBuild{})
}

func runAllMultiCmdTests(t *testing.T, cases map[string]processorTestCase, plugin manifest.Plugin) {
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			utils.RunCmdWithEnv = func(cmdline string, env map[string]string) error {
				for _, v := range test.expect["cmdline"].([]string) {
					if v == cmdline {
						return nil
					}
				}
				return fmt.Errorf("cmd: %v not found in %v", cmdline, test.expect["cmdline"].([]string))
			}

			if err := loadTestData(test.in, plugin); err != nil {
				t.Errorf("Error: %v", err)
				t.Fail()
			}
		})
	}
}
