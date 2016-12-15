package plugins

import (
	"fmt"
	"github.com/fatih/color"
	"testing"

	consul "github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func TestOutdated(t *testing.T) {
	runAllConsulTests(t,
		map[string]processorTestCase{
			"simple": {
				in: `---
consul-address: "127.0.0.1"
full-name: "test"`,
				expect: map[string]interface{}{
					"cmdline":  []string{""},
					"consulKV": []string{`services/outdated/test`},
				},
			},
		},
		Outdated{})
}

func runAllConsulTests(t *testing.T, cases map[string]processorTestCase, plugin manifest.Plugin) {
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

		utils.PutConsulKv = func(client *consul.Client, key string, value string) error {
			for _, v := range test.expect["consulKV"].([]string) {
				if v == key {
					return nil
				}
			}
			return fmt.Errorf("consulKV: %v not found in %v", key, test.expect["consulKV"].([]string))
		}

		utils.DelConsulKv = func(client *consul.Client, key string) error {
			return nil
		}

		if err := loadTestData(test.in, plugin); err == nil {
			color.Green("%v: Ok\n", name)
		} else {
			color.Red("error %v\n: failed!", err)
			t.Fail()
		}
	}
}
