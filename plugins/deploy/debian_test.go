package deploy

import (
	"fmt"
	"github.com/fatih/color"
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

func TestDeployDebian(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"install": {
				in: `---
parallel: 1
consul-address: "consul.test.ru"
cluster: "test.ru"
ssh-user: "test_user"
ci-tools-path: "/var/test"
app-name: "test/package"
package: "package"
version: "0.0.0"`,
				expect: map[string]interface{}{
					"cmdline": []string{"dig +short test.ru | sort | uniq | parallel --tag --line-buffer -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@{} \"sudo /var/test/debian-way/deploy.sh --package='package' --version='0.0.0'\""},
				},
			},
			"uninstall": {
				in: `---
parallel: 1
consul-address: "consul.test.ru"
cluster: "test.ru"
ssh-user: "test_user"
ci-tools-path: "/var/test"
app-name: "test/package"
package: "package"
version: "0.0.0"
purge: true`,
				expect: map[string]interface{}{
					"cmdline": []string{"dig +short test.ru | sort | uniq | parallel --tag --line-buffer -j 1 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@{} \"sudo apt-get purge package -y\""},
				},
			},
		},
		DeployDebian{})
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

		utils.RegisterPluginData = func(plugin string, packageName string, data string, consulAddress string) error {
			return nil
		}

		utils.DeletePluginData = func(plugin string, packageName string, consulAddress string) error {
			return nil
		}

		utils.RandomString = func(length int) string {
			return "RANDOM_NAME"
		}

		if err := loadTestData(test.in, plugin); err == nil {
			color.Green("%v: Ok\n", name)
		} else {
			color.Red("error %v\n: failed!", err)
			t.Fail()
		}
	}
}
