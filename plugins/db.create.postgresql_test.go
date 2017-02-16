package plugins

import (
	"testing"

	"github.com/servehub/serve/utils/tests"
)

func TestDBCreatePostgresql(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"create": {
				In: `---
				purge: false
				ssh-user: "test_user"
				target: "target_db_test"
				`,
				Expects: []string{
					"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres createdb -O postgres \"target_db_test\"\"",
				},
			},
			"create with source": {
				In: `---
					purge: false
					ssh-user: "test_user"
					source: "source_db_test"
					target: "target_db_test"
				`,
				Expects: []string{
					"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres createdb -O postgres \"target_db_test\" && sudo -Hu postgres pg_dump \"source_db_test\" | sudo -Hu postgres psql \"target_db_test\"\"",
				},
			},
			"drop": {
				In: `---
					purge: true
					ssh-user: "test_user"
					target: "target_db_test"
				`,
				Expects: []string{
					"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres psql -c \\\"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='target_db_test';\\\" && sudo -Hu postgres dropdb --if-exists \"target_db_test\"\"",
				},
			},
			"drop with source": {
				In: `---
					purge: true
					ssh-user: "test_user"
					source: "source_db_test"
					target: "target_db_test"
				`,
				Expects: []string{
					"ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@<nil> \"sudo -Hu postgres psql -c \\\"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='target_db_test';\\\" && sudo -Hu postgres dropdb --if-exists \"target_db_test\"\"",
				},
			},
		},
		DBCreatePostgresql{})
}
