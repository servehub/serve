package build

import (
	"testing"

	"github.com/servehub/serve/utils/tests"
)

func TestTarballBuild(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
					files: []
					registry-url: test.ru
				`,
				Expects: []string{
					"rm -rf ./tarball.tmp && mkdir ./tarball.tmp",
					"tar -zcf tarball.tar.gz -C ./tarball.tmp/ .",
					"curl -vsSf -XPUT -T tarball.tar.gz test.ru",
				},
			},
		},
		TarballBuild{})
}
