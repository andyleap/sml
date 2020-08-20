package sml

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestGrammar(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output interface{}
	}{
		{
			"String",
			`a b c`,
			"a b c",
		},
		{
			"List",
			`- a d
- b e
- c f`,
			[]interface{}{"a d", "b e", "c f"},
		},
		{
			"Map",
			`a: b
c: d`,
			map[string]interface{}{"a": "b", "c": "d"},
		},
		{
			"Map of List",
			`a:
  - b
  - c
d:
  - e
  - f`,
			map[string]interface{}{"a": []interface{}{"b", "c"}, "d": []interface{}{"e", "f"}},
		},
		{
			"List of Map",
			`- a: b
  c: d
- e: f
  g: h`,
			[]interface{}{
				map[string]interface{}{
					"a": "b",
					"c": "d",
				},
				map[string]interface{}{
					"e": "f",
					"g": "h",
				},
			},
		},
		{
			"List of List",
			`- - a
  - b
- - c
  - d`,
			[]interface{}{
				[]interface{}{
					"a",
					"b",
				},
				[]interface{}{
					"c",
					"d",
				},
			},
		},
		{
			"Map of Map",
			`a:
  b: c
  d: e
f:
  g: h
  i: j`,
			map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c",
					"d": "e",
				},
				"f": map[string]interface{}{
					"g": "h",
					"i": "j",
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.input)
			out, err := Decode(reader)
			if err != nil {
				t.Error(err)
			}
			if diff := deep.Equal(out, test.output); diff != nil {
				t.Logf("%#v", out)
				t.Error(diff)
			}
		})
	}
}
