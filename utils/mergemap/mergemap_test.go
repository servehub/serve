package mergemap

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {

	for _, tuple := range []struct {
		name     string
		src      string
		dst      string
		expected string
	}{
		{
			name:     `empty`,
			src:      `{}`,
			dst:      `{}`,
			expected: `{}`,
		},
		{
			name:     `merge`,
			src:      `{"b":2}`,
			dst:      `{"a":1}`,
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     `rewrite`,
			src:      `{"a":0}`,
			dst:      `{"a":1}`,
			expected: `{"a":0}`,
		},
		{
			name:     `merge dicts`,
			src:      `{"a":{       "y":2}}`,
			dst:      `{"a":{"x":1       }}`,
			expected: `{"a":{"x":1, "y":2}}`,
		},
		{
			name:     `rewrire dicts`,
			src:      `{"a":{"x":2}}`,
			dst:      `{"a":{"x":1}}`,
			expected: `{"a":{"x":2}}`,
		},
		{
			name:     `merge and rewrite dicts`,
			src:      `{"a":{       "y":7, "z":8}}`,
			dst:      `{"a":{"x":1, "y":2       }}`,
			expected: `{"a":{"x":1, "y":7, "z":8}}`,
		},
		{
			name:     `merge dict dicts`,
			src:      `{"1": { "b":1, "2": { "3": {         "b":3, "n":[1,2]} }        }}`,
			dst:      `{"1": {        "2": { "3": {"a":"A",        "n":"xxx"} }, "a":3 }}`,
			expected: `{"1": { "b":1, "2": { "3": {"a":"A", "b":3, "n":[1,2]} }, "a":3 }}`,
		},
		{
			name:     `merge dict dicts`,
			dst:      `{"data_resources": { "mysql_stream_mapping": { "dev":"3306", "A":"3306", "B":"3307", "C":"3308"}}}`,
			src:      `{"data_resources": { "mysql_stream_mapping": {               "A":"3308",                         "5": {"host":"asdf", "port":"3310"}}}}`,
			expected: `{"data_resources": { "mysql_stream_mapping": { "dev":"3306", "A":"3308", "B":"3307", "C":"3308", "5": {"host":"asdf", "port":"3310"}}}}`,
		},
		{
			name:     `merge dict and list`,
			dst:      `{"build": { "sh": {"version": "1.0.34"}, "zip": {"format": "tar.gz"} }}`,
			src:      `{"build": [ {"sh": "echo hello"}, {"sh": "echo hello2"}, {"sh": {"sh": "echo hello3", "other": "ok"}}, {"zip": {"target": "/tmp/" }}, {"zip": {"format": "new" }} ]}`,
			expected: `{"build": [ {"sh": {"sh": "echo hello", "version": "1.0.34"}}, {"sh": {"sh": "echo hello2", "version": "1.0.34"}}, {"sh": {"sh": "echo hello3", "version": "1.0.34", "other": "ok"}}, {"zip": {"target": "/tmp/", "format": "tar.gz"}} , {"zip": {"format": "new"}} ]}`,
		},
		{
			name:     `merge list dicts 2`,
			dst:      `{"build": [ {"debian": {"install-root": "/local/serve/tools"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "serve"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/serve/tools", "user": "serve"}} ] }`,
		},
		{
			name:     `merge list dicts 3`,
			dst:      `{"build": [ {"debian": {"install-root": "/local/serve/tools"}}, {"sh": {"version": "1.0.34"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "serve"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/serve/tools", "user": "serve"}} ] }`,
		},
		{
			name:     `merge list dicts 4`,
			dst:      `{"build": [ {"debian": {"install-root": "/local/serve/tools"}}, {"sh": {"version": "1.0.34"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "serve"}}, {"sh": {"version": "1.123"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/serve/tools", "user": "serve"}}, {"sh": {"version": "1.123"}} ] }`,
		},
		{
			name:     `merge list dicts 5`,
			dst:      `{"build": [ {"debian": {"install-root": "/local/serve/tools"}}, {"sh": {"version": "1.0.34"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "serve"}}, {"sh": {"format": "zip"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/serve/tools", "user": "serve"}}, {"sh": {"version": "1.0.34", "format": "zip"}} ] }`,
		},
	} {
		t.Run(tuple.name, func(t *testing.T) {
			var dst map[string]interface{}
			if err := json.Unmarshal([]byte(tuple.dst), &dst); err != nil {
				t.Error(err)
				t.Fail()
			}

			var src map[string]interface{}
			if err := json.Unmarshal([]byte(tuple.src), &src); err != nil {
				t.Error(err)
				t.Fail()
			}

			var expected map[string]interface{}
			if err := json.Unmarshal([]byte(tuple.expected), &expected); err != nil {
				t.Error(err)
				t.Fail()
			}

			got, _ := Merge(dst, src)
			assert(t, expected, got)
		})
	}

}

func assert(t *testing.T, expected, got map[string]interface{}) {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Error:\nexpected %v, \ngot %v", expected, got)
		t.Fail()
	}
}
