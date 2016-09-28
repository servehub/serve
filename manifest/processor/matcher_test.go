package processor

import "testing"

func TestMatcher(t *testing.T) {
	runAllProcessorTests(t, func() Processor { return Matcher{} }, map[string]processorTestCase{
		"simple match": {
			in: `
				{
					"vars": {
						"env": "qa"
					},
					"deploy": {
						"host ? {{ vars.env }}": {
							"qa": "qa-host.com",
							"live": "live-host.com"
						}
					}
				}
			`,
			expect: `{"deploy":{"host":"qa-host.com"},"vars":{"env":"qa"}}`,
		},

		"regexp value": {
			in: `
				{
					"vars": {
						"env": "qa-ru"
					},
					"deploy": {
						"host ? {{ vars.env }}": {
							"qa-.*": "qa-host.com",
							"live": "live-host.com"
						}
					}
				}
			`,
			expect: `{"deploy":{"host":"qa-host.com"},"vars":{"env":"qa-ru"}}`,
		},

		"default value": {
			in: `
				{
					"vars": {
						"env": "live-ru"
					},
					"deploy": {
						"host ? {{ vars.env }}": {
							"qa-.*": "qa-host.com",
							"live": "live-host.com",
							"*": "other"
						}
					}
				}
			`,
			expect: `{"deploy":{"host":"other"},"vars":{"env":"live-ru"}}`,
		},

		"reordering": {
			in: `
				{
					"vars": {
						"env": "live"
					},
					"deploy": {
						"host ? {{ vars.env }}": {
							"*": "other",
							"live": "live-host.com"
						}
					}
				}
			`,
			expect: `{"deploy":{"host":"live-host.com"},"vars":{"env":"live"}}`,
		},

		"reordering 2": {
			in: `
				{
					"vars": {
						"env": "live"
					},
					"deploy": {
						"host ? {{ vars.env }}": {
							"live": "live-host.com",
							"*": "other"
						}
					}
				}
			`,
			expect: `{"deploy":{"host":"live-host.com"},"vars":{"env":"live"}}`,
		},

		"ignore quotes": {
			in: `
				{
					"vars": {
						"env": "live"
					},
					"deploy": {
						"host ? \"{{ vars.env }}\"": {
							"live": "live-host.com"
						}
					}
				}
			`,
			expect: `{"deploy":{"host":"live-host.com"},"vars":{"env":"live"}}`,
		},

		"not found": {
			in: `
				{
					"vars": {
						"env": "live"
					},
					"deploy": {
						"host ? {{ vars.env }}": {
							"qa": "qa-host.com",
							"dev": "dev-host.com"
						}
					}
				}
			`,
			expect: `{"deploy":{},"vars":{"env":"live"}}`,
		},

		"cycling refs": {
			in: `
				{
					"vars": {
						"env": "{{ vars.branch }}",
						"branch": ""
					},
					"deploy": {
						"host ? {{ vars.env }}": {
							"*": "other-host.com",
							"": "empty-host.com"
						}
					}
				}
			`,
			expect: `{"deploy":{"host":"empty-host.com"},"vars":{"branch":"","env":"{{ vars.branch }}"}}`,
		},

		"cycling match": {
			in: `
				{
					"info": {
						"feature ? {{ info.input }}": {
							"*": "input-{{ info.input }}",
							"": "not feature"
						},
						"input ? {{ vars.branch }}": {
							"*": "featured-{{ vars.branch }}",
							"": ""
						}
					},
					"deploy": {
						"host ? {{ info.feature }}": {
							"*": "featured-host.com",
							"": "master-host.com"
						}
					},
					"vars": {
						"branch": "feature-superman"
					}
				}
			`,
			expect: `{"deploy":{"host":"featured-host.com"},"info":{"feature":"input-{{ info.input }}","input":"featured-{{ vars.branch }}"},"vars":{"branch":"feature-superman"}}`,
		},

		"cycling match 2": {
			in: `
				{
					"info": {
						"feature ? {{ info.input }}": {
							"*": "input-{{ info.input }}",
							"": "not feature"
						},
						"input ? {{ vars.branch }}": {
							"*": "featured-{{ vars.branch }}",
							"": ""
						}
					},
					"deploy": {
						"host ? {{ info.feature }}": {
							"*": "featured-host.com",
							"not feature": "master-host.com"
						}
					},
					"vars": {
						"branch": ""
					}
				}
			`,
			expect: `{"deploy":{"host":"master-host.com"},"info":{"feature":"not feature","input":""},"vars":{"branch":""}}`,
		},
	})
}
