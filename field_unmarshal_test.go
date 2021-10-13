package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBaseField_UnmarshalYAML(t *testing.T) {
	type testFieldStruct struct {
		BaseField

		Str     string   `yaml:"str"`
		StrPtr  *string  `yaml:"str_ptr"`
		BoolPtr *bool    `yaml:"bool_ptr"`
		Other   []string `rs:"other"`

		NestedStruct struct {
			BaseField

			StringMap map[string]string `yaml:"string_map"`
			Array     [5]interface{}    `yaml:"array"`
		} `yaml:"nested_struct"`
	}

	tests := []struct {
		name     string
		yaml     string
		expected interface{}
	}{
		{
			name: "basic",
			yaml: `str: bar`,
			expected: &testFieldStruct{
				BaseField: BaseField{
					unresolvedFields: nil,
				},
				Str: "bar",
			},
		},
		{
			name: "basic nil",
			yaml: `str: `,
			expected: &testFieldStruct{
				BaseField: BaseField{
					unresolvedFields: nil,
				},
				Str: "",
			},
		},
		{
			name: "basic ptr nil",
			yaml: `
str_ptr: null
bool_ptr: null
`,
			expected: &testFieldStruct{
				BaseField: BaseField{
					unresolvedFields: nil,
				},
				StrPtr:  nil,
				BoolPtr: nil,
			},
		},
		{
			name: "basic+renderer",
			yaml: `str@a: echo bar`,
			expected: &testFieldStruct{
				BaseField: BaseField{
					unresolvedFields: map[unresolvedFieldKey]*unresolvedFieldValue{
						{
							yamlKey: "str",
							suffix:  "a",
						}: {
							fieldName:  "Str",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{
								{
									scalarData: "echo bar",
								},
							},
							renderers: []*suffixSpec{{name: "a"}},
						},
					},
				},
				Str: "",
			},
		},
		{
			name: "catchAll+renderer",
			yaml: `{ other_field_1@a: foo, other_field_2@b: bar }`,
			expected: &testFieldStruct{
				BaseField: BaseField{
					catchOtherFields: map[string]struct{}{
						"other_field_1": {},
						"other_field_2": {},
					},
					catchOtherCache: nil,
					unresolvedFields: map[unresolvedFieldKey]*unresolvedFieldValue{
						{
							yamlKey: "other_field_1",
							suffix:  "a",
						}: {
							fieldName:  "Other",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{
								{
									mapData: map[string]*alterInterface{
										"other_field_1": {scalarData: "foo"},
									},
								},
							},
							renderers:         []*suffixSpec{{name: "a"}},
							isCatchOtherField: true,
						},
						{
							yamlKey: "other_field_2",
							suffix:  "b",
						}: {
							fieldName:  "Other",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{
								{
									mapData: map[string]*alterInterface{
										"other_field_2": {scalarData: "bar"},
									},
								},
							},
							renderers:         []*suffixSpec{{name: "b"}},
							isCatchOtherField: true,
						},
					},
				},
				// `Other` field should NOT be initialized
				// it will be initialized during resolving
				Other: nil,
			},
		},
		{
			name: "nested+renderer",
			// editorconfig-checker-disable
			yaml: `---
str@a: echo bar
nested_struct@b:
  string_map:
    c@d|e|f: e
  array@f:
  - "1"
  - "2"
  - "3"
  - "4"
  - '5'
`,
			// editorconfig-checker-enable
			expected: &testFieldStruct{
				BaseField: BaseField{
					unresolvedFields: map[unresolvedFieldKey]*unresolvedFieldValue{
						{
							yamlKey: "str",
							suffix:  "a",
						}: {
							fieldName:  "Str",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{
								{
									scalarData: "echo bar",
								},
							},
							renderers: []*suffixSpec{{name: "a"}},
						},
						{
							yamlKey: "nested_struct",
							suffix:  "b",
						}: {
							fieldName:  "NestedStruct",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{
								{
									mapData: map[string]*alterInterface{
										"string_map": {
											mapData: map[string]*alterInterface{
												"c@d|e|f": {scalarData: "e"},
											},
										},
										"array@f": {
											sliceData: []*alterInterface{
												{scalarData: "1"},
												{scalarData: "2"},
												{scalarData: "3"},
												{scalarData: "4"},
												{scalarData: "5"},
											},
										},
									},
								},
							},
							renderers: []*suffixSpec{{name: "b"}},
						},
					},
				},
				Str: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Init(&testFieldStruct{}, nil).(*testFieldStruct)
			assert.EqualValues(t, 1, out._initialized)

			if !assert.NoError(t, yaml.Unmarshal([]byte(test.yaml), out)) {
				return
			}

			out._initialized = 0
			out._parentValue = reflect.Value{}
			for k := range out.unresolvedFields {
				out.unresolvedFields[k].fieldValue = reflect.Value{}
			}

			assert.EqualValues(t, 1, out.NestedStruct._initialized)
			out.NestedStruct._initialized = 0
			out.NestedStruct._parentValue = reflect.Value{}
			for k := range out.NestedStruct.unresolvedFields {
				out.NestedStruct.unresolvedFields[k].fieldValue = reflect.Value{}
			}

			assert.EqualValues(t, test.expected, out)
		})
	}
}

func TestBaseField_UnmarshalYAML_Init(t *testing.T) {
	type Inner struct {
		BaseField

		Foo string `yaml:"foo"`

		DeepInner struct {
			BaseField

			Bar string `yaml:"bar"`
		} `yaml:"deep"`
	}

	rh := &testRenderingHandler{}

	t.Run("struct", func(t *testing.T) {
		type T struct {
			BaseField

			Foo Inner `yaml:"foo"`
		}

		out := Init(&T{}, nil).(*T)

		assert.NoError(t, yaml.Unmarshal([]byte(`foo: { foo: bar }`), out))
		assert.Equal(t, "bar", out.Foo.Foo)
		assert.EqualValues(t, 1, out.Foo.BaseField._initialized)

		out = Init(&T{}, nil).(*T)

		assert.NoError(t, yaml.Unmarshal([]byte(`foo@echo: { foo: rendered-bar }`), out))
		assert.Equal(t, "", out.Foo.Foo)
		assert.Len(t, out.BaseField.unresolvedFields, 1)
		assert.Len(t, out.Foo.BaseField.unresolvedFields, 0)
		assert.EqualValues(t, 1, out.Foo.BaseField._initialized)

		out.ResolveFields(rh, -1)

		assert.EqualValues(t, "rendered-bar", out.Foo.Foo)
	})

	t.Run("struct inline", func(t *testing.T) {
		type T struct {
			BaseField

			Foo Inner `yaml:",inline"`
		}

		out := Init(&T{}, nil).(*T)

		assert.NoError(t, yaml.Unmarshal([]byte(`foo: bar`), out))
		assert.Equal(t, "bar", out.Foo.Foo)
		assert.EqualValues(t, 1, out.Foo.BaseField._initialized)

		out = Init(&T{}, nil).(*T)

		assert.NoError(t, yaml.Unmarshal([]byte(`foo@echo: { foo: rendered-bar }`), out))
		assert.Equal(t, "", out.Foo.Foo)
		assert.EqualValues(t, 1, out.Foo.BaseField._initialized)
		assert.Len(t, out.BaseField.unresolvedFields, 0)
		assert.Len(t, out.Foo.BaseField.unresolvedFields, 1)
	})

	t.Run("struct embedded ", func(t *testing.T) {
		// nolint:unused
		type T struct {
			BaseField

			Inner `yaml:"inner"`
		}

		// TODO
	})

	t.Run("struct embedded inline", func(t *testing.T) {
		// nolint:unused
		type T struct {
			BaseField

			Inner `yaml:",inline"`
		}

		// TODO
	})

	t.Run("ptr", func(t *testing.T) {
		// nolint:unused
		type T struct {
			BaseField

			Foo *Inner `yaml:"foo"`
		}

		// TODO
	})

	t.Run("ptr inline", func(t *testing.T) {
		// nolint:unused
		type T struct {
			BaseField

			Foo *Inner `yaml:",inline"`
		}

		// TODO
	})

	t.Run("ptr embedded ", func(t *testing.T) {
		// nolint:unused
		type T struct {
			BaseField

			*Inner `yaml:"inner"`
		}

		// TODO
	})

	t.Run("ptr embedded inline", func(t *testing.T) {
		// nolint:unused
		type T struct {
			BaseField

			*Inner `yaml:",inline"`
		}

		// TODO
	})
}
