package release

import (
	"fmt"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func loadTestData(data string, plugin manifest.Plugin) error {
	if json, err := yaml.YAMLToJSON([]byte(data)); err != nil {
		return fmt.Errorf("Parser error: %v", err)
	} else {
		return plugin.Run(*manifest.LoadJSON(string(json)))
	}
}

type processorTestCase struct {
	in     string
	expect map[string]interface{}
}

func TestReleaseDebian(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"simple": {
				in: `---
cluster: "test.ru"
ssh-user: test_user
ci-tools-path: /var/test
package: package
site: test-site
mode: stage`,
				expect: map[string]interface{}{
					"cmdline": []string{"dig +short test.ru | sort | uniq | parallel --tag --line-buffer -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@{} \"sudo /var/test/debian-way/release.sh --package='package' --site='test-site' --mode='stage'\""},
				},
			},
		},
		ReleaseDebian{})
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
