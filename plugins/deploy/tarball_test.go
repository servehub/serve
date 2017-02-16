package deploy

import (
	"testing"

	"github.com/servehub/serve/utils/tests"
)

func TestDeployTarball(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"install": {
				In: `---
					cluster: "test.ru"
					ssh-user: "test_user"
					package-name: "test_name"
					package-uri: "test_name.ru"
					install-root: "/local/test/tarball"
					user: "user"
					group: "group"
					consul-address: "consul.test.ru"
					hooks: []
				`,
				Expects: []string{
					"dig +short test.ru | sort | uniq | parallel --tag --line-buffer -j 50 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@{} \"curl -vsSf -o /tmp/tarball-RANDOM_NAME.tar.gz test_name.ru && rm -rf /tmp/tarball-RANDOM_NAME/ && mkdir -p /tmp/tarball-RANDOM_NAME/ && tar xzf /tmp/tarball-RANDOM_NAME.tar.gz -C /tmp/tarball-RANDOM_NAME/ && sudo rm -rf /local/test/tarball/test_name && sudo mkdir -p /local/test/tarball/test_name && sudo mv /tmp/tarball-RANDOM_NAME/* /local/test/tarball/test_name/ && sudo chown -R user:group /local/test/tarball/test_name/ && rm -rf /tmp/tarball-RANDOM_NAME.tar.gz /tmp/tarball-RANDOM_NAME/\"",
				},
			},
			"install with hooks": {
				In: `---
					cluster: "test.ru"
					ssh-user: "test_user"
					package-name: "test_name"
					package-uri: "test_name.ru"
					install-root: "/local/test/tarball"
					user: "user"
					group: "group"
					consul-address: "consul.test.ru"
					hooks:
					  - postinstall: test1.sh
					  - postinstall: test2.sh
				`,
				Expects: []string{
					"dig +short test.ru | sort | uniq | parallel --tag --line-buffer -j 50 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@{} \"curl -vsSf -o /tmp/tarball-RANDOM_NAME.tar.gz test_name.ru && rm -rf /tmp/tarball-RANDOM_NAME/ && mkdir -p /tmp/tarball-RANDOM_NAME/ && tar xzf /tmp/tarball-RANDOM_NAME.tar.gz -C /tmp/tarball-RANDOM_NAME/ && sudo rm -rf /local/test/tarball/test_name && sudo mkdir -p /local/test/tarball/test_name && sudo mv /tmp/tarball-RANDOM_NAME/* /local/test/tarball/test_name/ && sudo chown -R user:group /local/test/tarball/test_name/ && rm -rf /tmp/tarball-RANDOM_NAME.tar.gz /tmp/tarball-RANDOM_NAME/ && sudo test1.sh && sudo test2.sh\"",
				},
			},
			"uninstall": {
				In: `---
					cluster: "test.ru"
					ssh-user: "test_user"
					package-name: "test_name"
					package-uri: "test_name.ru"
					install-root: "/local/test/tarball"
					user: "user"
					group: "group"
					consul-address: "consul.test.ru"
					hooks: []
					purge: true
				`,
				Expects: []string{
					"dig +short test.ru | sort | uniq | parallel --tag --line-buffer -j 50 ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null test_user@{} \"sudo rm -rf /local/test/tarball/test_name\"",
				},
			},
		},
		DeployTarball{})
}
