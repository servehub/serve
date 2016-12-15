package plugins

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

func TestDBCreatePostgresql(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"create": {
				in: `---
purge: false
ssh-user: "test_user"
target: "target_db_test"`,
				expect: map[string]interface{}{
					"cmdline": []string{"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres createdb -O postgres \"target_db_test\"\""},
				},
			},
			"create with source": {
				in: `---
purge: false
ssh-user: "test_user"
source: "source_db_test"
target: "target_db_test"`,
				expect: map[string]interface{}{
					"cmdline": []string{"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres createdb -O postgres \"target_db_test\" && sudo -Hu postgres pg_dump \"source_db_test\" | sudo -Hu postgres psql \"target_db_test\"\""},
				},
			},
			"drop": {
				in: `---
purge: true
ssh-user: "test_user"
target: "target_db_test"`,
				expect: map[string]interface{}{
					"cmdline": []string{"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres psql -c \\\"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='target_db_test';\\\" && sudo -Hu postgres dropdb --if-exists \"target_db_test\"\""},
				},
			},
			"drop with source": {
				in: `---
purge: true
ssh-user: "test_user"
source: "source_db_test"
target: "target_db_test"`,
				expect: map[string]interface{}{
					"cmdline": []string{"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres psql -c \\\"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='target_db_test';\\\" && sudo -Hu postgres dropdb --if-exists \"target_db_test\"\""},
				},
			},
		},
		DBCreatePostgresql{})
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
