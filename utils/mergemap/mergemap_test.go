package mergemap

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMerge(t *testing.T) {
	for _, tuple := range []struct {
		src      string
		dst      string
		expected string
	}{
		{
			src:      `{}`,
			dst:      `{}`,
			expected: `{}`,
		},
		{
			src:      `{"b":2}`,
			dst:      `{"a":1}`,
			expected: `{"a":1,"b":2}`,
		},
		{
			src:      `{"a":0}`,
			dst:      `{"a":1}`,
			expected: `{"a":0}`,
		},
		{
			src:      `{"a":{       "y":2}}`,
			dst:      `{"a":{"x":1       }}`,
			expected: `{"a":{"x":1, "y":2}}`,
		},
		{
			src:      `{"a":{"x":2}}`,
			dst:      `{"a":{"x":1}}`,
			expected: `{"a":{"x":2}}`,
		},
		{
			src:      `{"a":{       "y":7, "z":8}}`,
			dst:      `{"a":{"x":1, "y":2       }}`,
			expected: `{"a":{"x":1, "y":7, "z":8}}`,
		},
		{
			src:      `{"1": { "b":1, "2": { "3": {         "b":3, "n":[1,2]} }        }}`,
			dst:      `{"1": {        "2": { "3": {"a":"A",        "n":"xxx"} }, "a":3 }}`,
			expected: `{"1": { "b":1, "2": { "3": {"a":"A", "b":3, "n":[1,2]} }, "a":3 }}`,
		},
		{
			dst:      `{"data_resources": { "mysql_stream_mapping": { "dev":"3306", "A":"3306", "B":"3307", "C":"3308"}}}`,
			src:      `{"data_resources": { "mysql_stream_mapping": {               "A":"3308",                         "5": {"host":"asdf", "port":"3310"}}}}`,
			expected: `{"data_resources": { "mysql_stream_mapping": { "dev":"3306", "A":"3308", "B":"3307", "C":"3308", "5": {"host":"asdf", "port":"3310"}}}}`,
		},
		{
			dst:      `{"build": { "sh": {"version": "1.0.34"}, "zip": {"format": "tar.gz"} }}`,
			src:      `{"build": [ {"sh": "echo hello"}, {"sh": "echo hello2"}, {"sh": {"sh": "echo hello3", "other": "ok"}}, {"zip": {"target": "/tmp/" }}, {"zip": {"format": "new" }} ]}`,
			expected: `{"build": [ {"sh": {"sh": "echo hello", "version": "1.0.34"}}, {"sh": {"sh": "echo hello2", "version": "1.0.34"}}, {"sh": {"sh": "echo hello3", "version": "1.0.34", "other": "ok"}}, {"zip": {"target": "/tmp/", "format": "tar.gz"}} , {"zip": {"format": "new"}} ]}`,
		},
		{
			dst:      `{"build": [ {"debian": {"install-root": "/local/innova/tools"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "innova"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/innova/tools", "user": "innova"}} ] }`,
		},
		{
			dst:      `{"build": [ {"debian": {"install-root": "/local/innova/tools"}}, {"sh": {"version": "1.0.34"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "innova"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/innova/tools", "user": "innova"}} ] }`,
		},
		{
			dst:      `{"build": [ {"debian": {"install-root": "/local/innova/tools"}}, {"sh": {"version": "1.0.34"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "innova"}}, {"sh": {"version": "1.123"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/innova/tools", "user": "innova"}}, {"sh": {"version": "1.123"}} ] }`,
		},
		{
			dst:      `{"build": [ {"debian": {"install-root": "/local/innova/tools"}}, {"sh": {"version": "1.0.34"}} ] }`,
			src:      `{"build": [ {"debian": {"user": "innova"}}, {"sh": {"format": "zip"}} ] }`,
			expected: `{"build": [ {"debian": {"install-root": "/local/innova/tools", "user": "innova"}}, {"sh": {"version": "1.0.34", "format": "zip"}} ] }`,
		},
	} {
		var dst map[string]interface{}
		if err := json.Unmarshal([]byte(tuple.dst), &dst); err != nil {
			t.Error(err)
			continue
		}

		var src map[string]interface{}
		if err := json.Unmarshal([]byte(tuple.src), &src); err != nil {
			t.Error(err)
			continue
		}

		var expected map[string]interface{}
		if err := json.Unmarshal([]byte(tuple.expected), &expected); err != nil {
			t.Error(err)
			continue
		}

		got, _ := Merge(dst, src)
		assert(t, expected, got)
	}
}

func assert(t *testing.T, expected, got map[string]interface{}) {
	expectedBuf, err := json.Marshal(expected)
	if err != nil {
		t.Error(err)
		return
	}
	gotBuf, err := json.Marshal(got)
	if err != nil {
		t.Error(err)
		return
	}
	if bytes.Compare(expectedBuf, gotBuf) != 0 {
		t.Errorf("\nexpected %s, \ngot %s", string(expectedBuf), string(gotBuf))
		return
	}
}
