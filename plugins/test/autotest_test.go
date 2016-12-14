package test

import (
	"fmt"
	"testing"
	"github.com/fatih/color"

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

func TestAutotest_Run(t *testing.T) {
	runAllMultiCmdTests(t, map[string]processorTestCase{
		"simple": {
			in: `---
project: "test"
version: "0.0.0"
repo: "git@test.ru:test.git"
suite: "test-test"`,
			expect: map[string]interface{}{
				"cmdline": []string{"rm -rf tests && git clone --depth 1 --single-branch --recursive git@test.ru:test.git tests",
									"cd tests/ && ./test.sh --project=test --version=0.0.0 --suite=test-test"},
				},
		},
	},
	TestAutotest{})
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
