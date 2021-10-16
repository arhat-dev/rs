package rs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBaseField_MarshalYAML(t *testing.T) {

	valTrue := true
	valFalse := false
	pTrue := &valTrue
	pFalse := &valFalse

	type (
		Foo              struct{ Foo string }
		FooWithBaseField struct {
			BaseField
			Foo string
		}
	)

	tests := []struct {
		name string

		data     Field
		expected interface{}
	}{
		{
			name:     "Empty",
			data:     &struct{ BaseField }{},
			expected: &struct{}{},
		},
		{
			name: "String",
			data: &struct {
				BaseField
				Str string
			}{Str: "str"},
			expected: &struct{ Str string }{Str: "str"},
		},
		{
			name: "Int",
			data: &struct {
				BaseField
				Int int
			}{Int: 100},
			expected: &struct{ Int int }{Int: 100},
		},
		{
			name: "Float",
			data: &struct {
				BaseField
				Float float64
			}{Float: 10.1},
			expected: &struct{ Float float64 }{Float: 10.1},
		},
		{
			name: "Ptr with Value",
			data: &struct {
				BaseField
				Bool *bool
			}{Bool: pTrue},
			expected: &struct{ Bool *bool }{Bool: pTrue},
		},
		{
			name: "Ptr No Value",
			data: &struct {
				BaseField
				Bool *bool
			}{Bool: nil},
			expected: &struct{ Bool *bool }{Bool: nil},
		},
		{
			name: "Ptr No Value omitempty",
			data: &struct {
				BaseField
				Bool *bool `yaml:",omitempty"`
			}{Bool: nil},
			expected: &struct {
				Bool *bool `yaml:",omitempty"`
			}{Bool: nil},
		},
		{
			name: "Ptr with Value omitempty",
			data: &struct {
				BaseField
				Bool *bool `yaml:",omitempty"`
			}{Bool: pFalse},
			expected: &struct {
				Bool *bool `yaml:",omitempty"`
			}{Bool: pFalse},
		},
		{
			name: "Slice",
			data: &struct {
				BaseField
				Bool []bool
			}{Bool: nil},
			expected: &struct {
				Bool []bool
			}{Bool: nil},
		},
		{
			name: "Slice omitempty",
			data: &struct {
				BaseField
				Slice []bool `yaml:",omitempty"`
			}{Slice: nil},
			expected: &struct {
				Slice []bool `yaml:",omitempty"`
			}{Slice: nil},
		},
		{
			// to address https://github.com/go-yaml/yaml/issues/362
			name: "Inline",
			data: &struct {
				BaseField
				Struct struct {
					BaseField
					Foo string
				} `yaml:",inline"`
			}{Struct: struct {
				BaseField
				Foo string
			}{Foo: "foo"}},
			expected: &struct {
				Struct struct {
					Foo string
				} `yaml:",inline"`
			}{Struct: struct{ Foo string }{Foo: "foo"}},
		},
		{
			// to address https://github.com/go-yaml/yaml/issues/362
			name: "Embedded Inline",
			data: &struct {
				BaseField
				FooWithBaseField `yaml:",inline"`
			}{FooWithBaseField: FooWithBaseField{Foo: "foo"}},
			expected: &struct {
				Foo `yaml:",inline"`
			}{Foo: Foo{Foo: "foo"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ret, err := yaml.Marshal(Init(test.data, nil))
			assert.NoError(t, err)

			expected, err := yaml.Marshal(test.expected)
			assert.NoError(t, err)

			t.Log(string(ret))
			assert.EqualValues(t, string(expected), string(ret))
		})
	}
}
