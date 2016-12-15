package build

import (
	"fmt"
	"github.com/fatih/color"
	"testing"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
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
	color.NoColor = false

	for name, test := range cases {
		utils.RunCmdWithEnv = func(cmdline string, env map[string]string) error {
			for _, v := range test.expect["cmdline"].([]string) {
				if v == cmdline {
					return nil
				}
			}
			return fmt.Errorf("cmd: %v not found in %v", cmdline, test.expect["cmdline"].([]string))
		}

		if err := loadTestData(test.in, plugin); err == nil {
			color.Green("%v: Ok\n", name)
		} else {
			color.Red("error %v\n: failed!", err)
			t.Fail()
		}
	}
}
