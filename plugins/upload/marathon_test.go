package upload

import (
	"testing"

	"github.com/servehub/serve/utils/tests"
)

func TestUploadMarathon(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
					unstable-url: "http://unstable.test.ru"
					stable-url: "http://stable.test.ru"
				`,
				Expects: []string{
					"curl -vsSf -o marathon.tar.gz http://unstable.test.ru",
					"curl -vsSf -XPUT -T marathon.tar.gz http://stable.test.ru",
				},
			},
		},
		UploadMarathon{})
}
