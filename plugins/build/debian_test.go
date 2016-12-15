package build

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

func TestBuildDebian(t *testing.T) {
	runAllDebianTests(t,
		map[string]processorTestCase{
			"empty depends": {
				in: `---
name: "test-package"
description: "test package"
version: "0.0.0"
build-number: "0"
stage-counter: "1"
package: "package"
category: "test"
maintainer-email: "bamboo@inn.ru"
maintainer-name: "Continuous Integration"
install-root: "/local/test"
ci-tools-path: "/var/go/inn-ci-tools"
distribution: "unstable"
init: "debian-way"
consul-supervisor: true
daemon: ""
daemon-args: ""
daemon-port: ""
make-pidfile: "yes"
service-owner: "tester"
daemon-user: "tester"
depends: ""
cron: ""`,
				expect: map[string]interface{}{
					"cmdline": "/var/go/inn-ci-tools/go/debian-build.sh --distribution=unstable",
					"env": map[string]string{
						"MANIFEST_PACKAGE":                       "package",
						"MANIFEST_INFO_NAME":                     "test-package",
						"MANIFEST_INFO_VERSION":                  "0.0.0",
						"MANIFEST_BUILD_DEBIAN_SECTION":          "test",
						"MANIFEST_INFO_CATEGORY":                 "test",
						"MANIFEST_BUILD_DEBIAN_MAINTAINER_NAME":  "Continuous Integration",
						"MANIFEST_BUILD_DEBIAN_MAINTAINER_EMAIL": "bamboo@inn.ru",
						"MANIFEST_BUILD_DEBIAN_INSTALL_ROOT":     "/local/test",
						"MANIFEST_BUILD_DEBIAN_DAEMON":           "",
						"MANIFEST_BUILD_DEBIAN_DAEMON_ARGS":      "",
						"MANIFEST_BUILD_DEBIAN_SERVICE_OWNER":    "tester",
						"MANIFEST_BUILD_DEBIAN_DAEMON_USER":      "tester",
						"MANIFEST_BUILD_DEBIAN_DAEMON_PORT":      "",
						"MANIFEST_BUILD_DEBIAN_MAKE_PIDFILE":     "yes",
						"MANIFEST_BUILD_DEBIAN_DEPENDS":          "",
						"MANIFEST_BUILD_DEBIAN_DESCRIPTION":      "test package",
						"MANIFEST_BUILD_DEBIAN_INIT":             "debian-way",
						"MANIFEST_BUILD_DEBIAN_CRON":             "",
						"GO_PIPELINE_LABEL":                      "0",
						"GO_STAGE_COUNTER":                       "1",
					}},
			},
			"list depends": {
				in: `---
name: "test-package"
description: "test package"
version: "0.0.0"
build-number: "0"
stage-counter: "1"
package: "package"
category: "test"
maintainer-email: "bamboo@inn.ru"
maintainer-name: "Continuous Integration"
install-root: "/local/test"
ci-tools-path: "/var/go/inn-ci-tools"
distribution: "unstable"
init: "debian-way"
consul-supervisor: true
daemon: ""
daemon-args: ""
daemon-port: ""
make-pidfile: "yes"
service-owner: "tester"
daemon-user: "tester"
depends:
  - python
  - go
cron: ""`,
				expect: map[string]interface{}{
					"cmdline": "/var/go/inn-ci-tools/go/debian-build.sh --distribution=unstable",
					"env": map[string]string{
						"MANIFEST_PACKAGE":                       "package",
						"MANIFEST_INFO_NAME":                     "test-package",
						"MANIFEST_INFO_VERSION":                  "0.0.0",
						"MANIFEST_BUILD_DEBIAN_SECTION":          "test",
						"MANIFEST_INFO_CATEGORY":                 "test",
						"MANIFEST_BUILD_DEBIAN_MAINTAINER_NAME":  "Continuous Integration",
						"MANIFEST_BUILD_DEBIAN_MAINTAINER_EMAIL": "bamboo@inn.ru",
						"MANIFEST_BUILD_DEBIAN_INSTALL_ROOT":     "/local/test",
						"MANIFEST_BUILD_DEBIAN_DAEMON":           "",
						"MANIFEST_BUILD_DEBIAN_DAEMON_ARGS":      "",
						"MANIFEST_BUILD_DEBIAN_SERVICE_OWNER":    "tester",
						"MANIFEST_BUILD_DEBIAN_DAEMON_USER":      "tester",
						"MANIFEST_BUILD_DEBIAN_DAEMON_PORT":      "",
						"MANIFEST_BUILD_DEBIAN_MAKE_PIDFILE":     "yes",
						"MANIFEST_BUILD_DEBIAN_DEPENDS":          "python, go",
						"MANIFEST_BUILD_DEBIAN_DESCRIPTION":      "test package",
						"MANIFEST_BUILD_DEBIAN_INIT":             "debian-way",
						"MANIFEST_BUILD_DEBIAN_CRON":             "",
						"GO_PIPELINE_LABEL":                      "0",
						"GO_STAGE_COUNTER":                       "1",
					}},
			},
			"string depends": {
				in: `---
name: "test-package"
description: "test package"
version: "0.0.0"
build-number: "0"
stage-counter: "1"
package: "package"
category: "test"
maintainer-email: "bamboo@inn.ru"
maintainer-name: "Continuous Integration"
install-root: "/local/test"
ci-tools-path: "/var/go/inn-ci-tools"
distribution: "unstable"
init: "debian-way"
consul-supervisor: true
daemon: ""
daemon-args: ""
daemon-port: ""
make-pidfile: "yes"
service-owner: "tester"
daemon-user: "tester"
depends: "python, go"
cron: ""`,
				expect: map[string]interface{}{
					"cmdline": "/var/go/inn-ci-tools/go/debian-build.sh --distribution=unstable",
					"env": map[string]string{
						"MANIFEST_PACKAGE":                       "package",
						"MANIFEST_INFO_NAME":                     "test-package",
						"MANIFEST_INFO_VERSION":                  "0.0.0",
						"MANIFEST_BUILD_DEBIAN_SECTION":          "test",
						"MANIFEST_INFO_CATEGORY":                 "test",
						"MANIFEST_BUILD_DEBIAN_MAINTAINER_NAME":  "Continuous Integration",
						"MANIFEST_BUILD_DEBIAN_MAINTAINER_EMAIL": "bamboo@inn.ru",
						"MANIFEST_BUILD_DEBIAN_INSTALL_ROOT":     "/local/test",
						"MANIFEST_BUILD_DEBIAN_DAEMON":           "",
						"MANIFEST_BUILD_DEBIAN_DAEMON_ARGS":      "",
						"MANIFEST_BUILD_DEBIAN_SERVICE_OWNER":    "tester",
						"MANIFEST_BUILD_DEBIAN_DAEMON_USER":      "tester",
						"MANIFEST_BUILD_DEBIAN_DAEMON_PORT":      "",
						"MANIFEST_BUILD_DEBIAN_MAKE_PIDFILE":     "yes",
						"MANIFEST_BUILD_DEBIAN_DEPENDS":          "python, go",
						"MANIFEST_BUILD_DEBIAN_DESCRIPTION":      "test package",
						"MANIFEST_BUILD_DEBIAN_INIT":             "debian-way",
						"MANIFEST_BUILD_DEBIAN_CRON":             "",
						"GO_PIPELINE_LABEL":                      "0",
						"GO_STAGE_COUNTER":                       "1",
					}},
			},
		},
		BuildDebian{})
}

func runAllDebianTests(t *testing.T, cases map[string]processorTestCase, plugin manifest.Plugin) {
	color.NoColor = false

	for name, test := range cases {
		utils.RunCmdWithEnv = func(cmdline string, env map[string]string) error {
			if cmdline != test.expect["cmdline"] {
				return fmt.Errorf("%v != %v", cmdline, test.expect["cmdline"])
			}
			tv := test.expect["env"].(map[string]string)
			for k, v := range env {
				if v != tv[k] {
					return fmt.Errorf("%v: %v != %v", k, v, tv[k])
				}
			}
			if len(env) != len(tv) {
				return fmt.Errorf("env len error: %v != %v", len(env), len(tv))
			}
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
