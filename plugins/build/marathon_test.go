package build

import (
	"testing"

	"github.com/servehub/serve/utils/tests"
)

func TestMarathonBuild(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
					source: target/pack
					registry-url: test.ru
				`,
				Expects: []string{
					"tar -zcf marathon.tar.gz -C target/pack/ .",
					"curl -vsSf -XPUT -T marathon.tar.gz test.ru",
				},
			},
		},
		MarathonBuild{})
}
