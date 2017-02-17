package upload

import (
	"testing"

	"github.com/servehub/serve/tests"
)

func TestUploadTarball(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
					unstable-url: "http://unstable.test.ru"
					stable-url: "http://stable.test.ru"
				`,
				Expects: []string{
					"curl -vsSf -o tarball.tar.gz http://unstable.test.ru",
					"curl -vsSf -XPUT -T tarball.tar.gz http://stable.test.ru",
				},
			},
		},
		UploadTarball{})
}
