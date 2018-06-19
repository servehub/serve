package tests

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

type TestCase struct {
	In     string
	Expects []string
}

func RunAllMultiCmdTests(t *testing.T, cases map[string]TestCase, plugin manifest.Plugin) {
	utils.RegisterPluginData = func(plugin string, packageName string, data string, consulAddress string) error {
		return nil
	}

	utils.DeletePluginData = func(plugin string, packageName string, consulAddress string) error {
		return nil
	}

	utils.RandomString = func(length uint) string {
		return "RANDOM_NAME"
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			var number int32 = 0

			utils.RunCmdWithEnv = func(cmdline string, env map[string]string) error {
				if int(number) < len(test.Expects) && cmdline == test.Expects[number] {
					atomic.AddInt32(&number, 1)
					return nil
				}

				return fmt.Errorf("\ncmd: \ngiven: %v \nexpected: %v", cmdline, test.Expects[number])
			}

			if err := loadTestData(utils.StripLeftMargin(test.In), plugin); err != nil {
				t.Errorf("Error: %v", err)
				t.Fail()
			}
		})
	}
}

func loadTestData(data string, plugin manifest.Plugin) error {
	if json, err := yaml.YAMLToJSON([]byte(data)); err != nil {
		return fmt.Errorf("Parser error: %v", err)
	} else {
		return plugin.Run(*manifest.ParseJSON(string(json)))
	}
}
