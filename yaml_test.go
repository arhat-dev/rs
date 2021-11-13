package rs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestIsEmpty(t *testing.T) {
	for _, test := range []struct {
		input string
		empty bool
	}{
		{"map: val", false},
		{"1.1", false},
		{`"foo"`, false},
		{"[]", false},

		{"", true},
		{"# all comments", true},
		{"null", true},
		{"~", true},
	} {
		n := new(yaml.Node)
		assert.NoError(t, yaml.Unmarshal([]byte(test.input), n))
		assert.Equal(t, test.empty, isEmpty(n), "test data: %q", test.input)
	}
}

func TestCheckScalarType(t *testing.T) {
	// editorconfig-checker-disable
	const _input = `
str_double_quoted: "str"
str_single_quoted: 'str'
str_literal: |
  str
str_folded: >
  str
int: 1
float: 1.1
bool: true
null_implicit:
null_explicit: null
not_null: {}

non_scalar:
  foo: bar
`
	// editorconfig-checker-enable
	n := new(yaml.Node)
	assert.NoError(t, yaml.Unmarshal([]byte(_input), n))
	n = prepareYamlNode(n)
	mkv, err := unmarshalYamlMap(n.Content)
	assert.NoError(t, err)

	input := make(map[string]*yaml.Node)
	for _, v := range mkv {
		input[v[0].Value] = v[1]
	}

	for _, test := range []struct {
		key      string
		tag      string
		expected bool
	}{
		{"str_double_quoted", strTag, true},
		{"str_single_quoted", strTag, true},
		{"str_literal", strTag, true},
		{"str_folded", strTag, true},
		{"int", intTag, true},
		{"float", floatTag, true},
		{"bool", boolTag, true},
		{"null_implicit", nullTag, true},
		{"null_explicit", nullTag, true},
		{"not_null", nullTag, false},

		{"non_scalar", strTag, false},
		{"non_scalar", intTag, false},
		{"non_scalar", floatTag, false},
		{"non_scalar", boolTag, false},
		{"non_scalar", nullTag, false},
	} {
		t.Run(test.key, func(t *testing.T) {
			n = prepareYamlNode(n)
			assert.Equal(t, test.expected, checkScalarType(input[test.key], test.tag))
		})
	}
}
