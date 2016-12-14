package plugins

import (
	"fmt"
	"testing"
	"github.com/fatih/color"
	"time"

	consul "github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)


func TestOutdated(t *testing.T) {
	runAllConsulTests(t, map[string]processorTestCase{
		"simple": {
			in: `---
consul-address: "127.0.0.1"
full-name: "test"`,
			expect: map[string]interface{}{
				"cmdline": []string{""},
				"consulKV": []string{fmt.Sprintf(`services/outdated/test={"endOfLife":%d}`, time.Now().Add(0).UnixNano()/int64(time.Millisecond))},
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
				if v == fmt.Sprintf("%v=%v", key, value) {
					return nil
				}
			}
			return fmt.Errorf("consulKV: %v=%v not found in %v", key, value, test.expect["consulKV"].([]string))
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
