package test

import (
	"testing"

	"github.com/servehub/serve/utils/tests"
)

func TestTestAutotest(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
					project: "test"
					version: "0.0.0"
					repo: "git@test.ru:test.git"
					suite: "test-test"
					environment: {}
				`,
				Expects: []string{
					"rm -rf autotest && git clone --depth 1 --single-branch --recursive git@test.ru:test.git autotest",
					"cd autotest/ && ./test.sh --project=test --version=0.0.0 --suite=test-test",
				},
			},
		},
		TestAutotest{})
}
