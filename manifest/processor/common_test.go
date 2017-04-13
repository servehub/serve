package processor

import (
	"testing"
	"github.com/servehub/utils/gabs"
	"fmt"
)

func TestAllProcessors(t *testing.T) {
	runAllProcessorTests(t, func() Processor { return allProcessors{} }, map[string]processorTestCase{
		"process all": {
			in: `
				{
					"info": {
						"feature ? {{ vars.branch }}": {
				      "": "",
				      "master": "",
				      "feature-(?P<feature>.+)": "{{ match.feature | lower | replace('\\W', '-') }}",
				      "*": "{{ vars.branch | lower | replace('\\W', '-') }}"
				    },
				    "feature-suffix ? {{ info.feature }}": {
				      "": "",
				      "*": "-{{ info.feature }}"
	          }
          },
			    "vars": {
			      "branch": "some-Test-branch/new"
			    }
				}
			`,
			expect: `{"info":{"feature":"some-test-branch-new","feature-suffix":"-some-test-branch-new"},"vars":{"branch":"some-Test-branch/new"}}`,
		},

		"modify composite variables": {
			in: `
				{
			    "info": {
						"name": "testov-test{{ vars.branch }}",
						"category": "web",
						"full-name": "{{ info.category }}/{{ info.name }}"
          },
          "vars": { "branch": "superman" },
          "kube": {
			      "name": "{{ info.full-name | replace('\\W|_', '-') }}"
			    }
				}
			`,
			expect: `{"info":{"category":"web","full-name":"web/testov-testsuperman","name":"testov-testsuperman"},"kube":{"name":"web-testov-testsuperman"},"vars":{"branch":"superman"}}`,
		},
	})
}


type allProcessors struct{}

func (a allProcessors) Process(tree *gabs.Container) error {
	procs := []Processor{
		Matcher{},
		Templater{},
	}

	for _, proc := range procs {
		if err := proc.Process(tree); err != nil {
			return fmt.Errorf("%T: %v", proc, err)
		}
	}
	return nil
}
