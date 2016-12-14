package upload

import (
	"testing"
)

func TestUploadTarball(t *testing.T) {
	runAllMultiCmdTests(t, map[string]processorTestCase{
		"simple": {
			in: `---
unstable-url: "http://unstable.test.ru"
stable-url: "http://stable.test.ru"`,
			expect: map[string]interface{}{
				"cmdline": []string{"curl -vsSf -o tarball.tar.gz http://unstable.test.ru",
								    "curl -vsSf -XPUT -T tarball.tar.gz http://stable.test.ru"},
				},
		},
	},
	UploadTarball{})
}
