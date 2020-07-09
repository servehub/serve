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
			in:     "",
			expect: `{"deploy":{"env":"qa"},"vars":{"env":"qa"}}`,
		},
		"empty": {
			yaml: `---
vars:
  env: qa
deploy:
  env: live`,
			in:     "",
			expect: `{"deploy":{"env":"live"},"vars":{"env":"qa"}}`,
		},
		"chain": {
			yaml: `---
vars: &v
  env: qa
  test: 1
deploy: &d
  <<: *v
  new-data: 2
release:
  <<: *d
  env: live`,
			in:     "",
			expect: `{"deploy":{"env":"qa","new-data":2,"test":1},"release":{"env":"live","new-data":2,"test":1},"vars":{"env":"qa","test":1}}`,
		},
	},
	)
}
