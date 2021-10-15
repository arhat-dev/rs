package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBaseField_UnmarshalYAML(t *testing.T) {
	type inlineStructWithBaseField struct {
		BaseField

		StringMap map[string]string `yaml:"string_map"`
		Array     [5]interface{}    `yaml:"array"`
	}

	type testFieldStruct struct {
		BaseField

		Str     string            `yaml:"str"`
		StrPtr  *string           `yaml:"str_ptr"`
		BoolPtr *bool             `yaml:"bool_ptr"`
		Other   map[string]string `rs:"other"`

		InlineWithBaseField inlineStructWithBaseField `yaml:",inline"`
	}

	tests := []struct {
		name string
		yaml string

		expectedUnmarshaled interface{}
		expectedResolved    interface{}

		expectUnmarshalErr bool
	}{
		{
			name: "basic",
			yaml: `str: bar`,

			expectedResolved: &testFieldStruct{Str: "bar"},
			expectedUnmarshaled: &testFieldStruct{
				BaseField: BaseField{unresolvedFields: nil},
				Str:       "bar",
			},
		},
		{
			name: "basic nil",
			yaml: `str: `,

			expectedResolved: &testFieldStruct{},
			expectedUnmarshaled: &testFieldStruct{
				BaseField: BaseField{unresolvedFields: nil},
				Str:       "",
			},
		},
		{
			name: "basic ptr nil",
			yaml: `
str_ptr: null
bool_ptr: null
`,
			expectedResolved: &testFieldStruct{},
			expectedUnmarshaled: &testFieldStruct{
				BaseField: BaseField{unresolvedFields: nil},
				StrPtr:    nil,
				BoolPtr:   nil,
			},
		},
		{
			name: "basic+renderer",
			yaml: `str@add-suffix-test: bar`,

			expectedResolved: &testFieldStruct{Str: "bar-test"},
			expectedUnmarshaled: &testFieldStruct{
				BaseField: BaseField{
					unresolvedFields: map[unresolvedFieldKey]*unresolvedFieldValue{
						{yamlKey: "str"}: {
							fieldName:   "Str",
							fieldValue:  reflect.Value{},
							rawDataList: []*alterInterface{{scalarData: "bar"}},
							renderers:   []*suffixSpec{{name: "add-suffix-test"}},
						},
					},
				},
				Str: "",
			},
		},
		{
			name: "catchAll same yaml key + renderer",
			yaml: `{other@echo: foo, other@echo|echo: bar}`,

			expectUnmarshalErr: true,
		},
		{
			name: "catchAll different yaml key + renderer",
			yaml: `{ other_field_1@echo: foo, other_field_2@add-suffix-test: bar }`,

			expectedResolved: &testFieldStruct{
				Other: map[string]string{
					"other_field_1": "foo",
					"other_field_2": "bar-test",
				},
			},
			expectedUnmarshaled: &testFieldStruct{
				BaseField: BaseField{
					catchOtherFields: map[string]struct{}{
						"other_field_1": {},
						"other_field_2": {},
					},
					catchOtherCache: nil,
					unresolvedFields: map[unresolvedFieldKey]*unresolvedFieldValue{
						{yamlKey: "other_field_1"}: {
							fieldName:  "Other",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{
								{
									mapData: map[string]*alterInterface{
										"other_field_1": {scalarData: "foo"},
									},
								},
							},
							renderers:         []*suffixSpec{{name: "echo"}},
							isCatchOtherField: true,
						},
						{yamlKey: "other_field_2"}: {
							fieldName:  "Other",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{{
								mapData: map[string]*alterInterface{
									"other_field_2": {scalarData: "bar"},
								},
							}},
							renderers:         []*suffixSpec{{name: "add-suffix-test"}},
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

string_map@echo|echo: |-
  c: e
array@echo|echo|echo: |-
  - "1"
  - "2"
  - "3"
  - "4"
  - '5'
`,
			// editorconfig-checker-enable

			expectedResolved: &testFieldStruct{
				InlineWithBaseField: inlineStructWithBaseField{
					StringMap: map[string]string{"c": "e"},
					Array:     [5]interface{}{"1", "2", "3", "4", "5"},
				},
			},
			expectedUnmarshaled: &testFieldStruct{
				InlineWithBaseField: inlineStructWithBaseField{
					BaseField: BaseField{
						unresolvedFields: map[unresolvedFieldKey]*unresolvedFieldValue{
							{yamlKey: "string_map"}: {
								fieldName:  "StringMap",
								fieldValue: reflect.Value{},
								rawDataList: []*alterInterface{{
									scalarData: "c: e",
								}},
								renderers: []*suffixSpec{
									{name: "echo"},
									{name: "echo"},
								},
							},
							{yamlKey: "array"}: {
								fieldName:  "Array",
								fieldValue: reflect.Value{},
								rawDataList: []*alterInterface{{
									scalarData: `- "1"
- "2"
- "3"
- "4"
- '5'`,
								}},
								renderers: []*suffixSpec{
									{name: "echo"},
									{name: "echo"},
									{name: "echo"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Run("unmarshal", func(t *testing.T) {
				out := Init(&testFieldStruct{}, nil).(*testFieldStruct)
				assert.EqualValues(t, 1, out._initialized)

				err := yaml.Unmarshal([]byte(test.yaml), out)
				if test.expectUnmarshalErr {
					assert.Error(t, err)
					return
				}

				if !assert.NoError(t, err) {
					return
				}

				// reset for assertion
				out._initialized = 0
				out._parentValue = reflect.Value{}
				for k := range out.unresolvedFields {
					out.unresolvedFields[k].fieldValue = reflect.Value{}
				}

				assert.EqualValues(t, 1, out.InlineWithBaseField._initialized)
				out.InlineWithBaseField._initialized = 0
				out.InlineWithBaseField._parentValue = reflect.Value{}
				for k := range out.InlineWithBaseField.unresolvedFields {
					out.InlineWithBaseField.unresolvedFields[k].fieldValue = reflect.Value{}
				}

				assert.EqualValues(t, test.expectedUnmarshaled, out)
			})

			if test.expectUnmarshalErr {
				return
			}

			t.Run("resolve", func(t *testing.T) {
				out := Init(&testFieldStruct{}, nil).(*testFieldStruct)
				assert.EqualValues(t, 1, out._initialized)

				if !assert.NoError(t, yaml.Unmarshal([]byte(test.yaml), out)) {
					return
				}

				assert.NoError(t, out.ResolveFields(&testRenderingHandler{}, -1))

				// reset for assertion
				out.BaseField = BaseField{}
				out.InlineWithBaseField.BaseField = BaseField{}

				assert.EqualValues(t, test.expectedResolved, out)
			})
		})
	}
}

func TestBaseField_UnmarshalYAML_Mixed_CatchOther(t *testing.T) {
	type Foo struct {
		BaseField

		Data map[string]string `rs:"other"`
	}

	a := Init(&Foo{}, nil).(*Foo)
	assert.NoError(t, yaml.Unmarshal([]byte(`{ a: a, b@echo: b, c: c }`), a))
	assert.EqualValues(t, map[string]string{
		"a": "a",
		"c": "c",
	}, a.Data)

	assert.NoError(t, a.ResolveFields(&testRenderingHandler{}, -1))
	assert.EqualValues(t, map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
	}, a.Data)
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
