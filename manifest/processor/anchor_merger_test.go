package processor

import (
	"testing"
)

func TestAnchorMerger(t *testing.T) {
	runAllProcessorTests(t, func() Processor { return AnchorMerger{} }, map[string]processorTestCase{
		"simple": {
			yaml: `---
vars: &v
  env: qa
deploy:
  <<: *v`,
			in: "",
			expect: `{"deploy":{"env":"qa"},"vars":{"env":"qa"}}`,
		},
		"empty": {
			yaml: `---
vars:
  env: qa
deploy:
  env: live`,
			in: "",
			expect: `{"deploy":{"env":"live"},"vars":{"env":"qa"}}`,
		},
		"chain": {
			yaml: `---
vars: &v
  env: qa
deploy: &d
  <<: *v
release:
  <<: *d`,
			in: "",
			expect: `{"deploy":{"env":"qa"},"release":{"env":"qa"},"vars":{"env":"qa"}}`,
		},

	},
	)
}
