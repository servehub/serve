package build

import (
	"testing"
)

func TestTarballBuild(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"simple": {
				in: `---
files: []
registry-url: test.ru`,
				expect: map[string]interface{}{
					"cmdline": []string{"rm -rf ./tarball.tmp && mkdir ./tarball.tmp",
						"tar -zcf tarball.tar.gz -C ./tarball.tmp/ .",
						"curl -vsSf -XPUT -T tarball.tar.gz test.ru",
					},
				},
			},
		},
		TarballBuild{})
}
