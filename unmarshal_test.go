package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBaseField_UnmarshalYAML_NoRS(t *testing.T) {
	tests := []struct {
		name string

		input string
		data  Field

		equivalent interface{}

		allowUnknown bool
		expectErr    bool
	}{
		{
			name:  "Simple No Tag Name",
			input: `a: b`,
			data: &struct {
				BaseField
				A string
			}{},
			equivalent: &struct {
				A string
			}{},
		},
		{
			name:  "Simple With Tag Name",
			input: `foo: b`,
			data: &struct {
				BaseField
				A string `yaml:"foo"`
			}{},
			equivalent: &struct {
				A string `yaml:"foo"`
			}{},
		},
		{
			name:         "Disallow Unknown Field",
			input:        `foo: b`,
			allowUnknown: false,
			expectErr:    true,

			data: &struct {
				BaseField

				A string
			}{},
		},
		{
			name:         "Allow Unknown Field",
			input:        `foo: b`,
			allowUnknown: true,
			expectErr:    false,

			data: &struct {
				BaseField
				A string
			}{},
			equivalent: &struct{ A string }{},
		},
		{
			name:  "With @ In Tag Name",
			input: `a@b: c`,
			data: &struct {
				BaseField

				A string `yaml:"a@b"`
			}{},

			equivalent: &struct {
				A string `yaml:"a@b"`
			}{},
		},
		{
			name: "Catch Other",

			input: `{ a: b, c: d }`,
			data: &struct {
				BaseField

				Data map[string]string `rs:"other"`
			}{},

			equivalent: &map[string]string{},
		},
		{
			name: "Catch Other Behave Like Inline Map",

			input: `{ a: b, c: d }`,
			data: &struct {
				BaseField

				Data map[string]string `rs:"other"`
			}{},

			equivalent: &struct {
				Data map[string]string `yaml:",inline"`
			}{},
		},
		{
			name: "Catch Other Indicated As Inline Map",

			input: `{ a: b, c: d }`,
			data: &struct {
				BaseField

				Data map[string]string `yaml:",inline"`
			}{},

			equivalent: &struct {
				Data map[string]string `yaml:",inline"`
			}{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := yaml.Unmarshal([]byte(test.input), Init(test.data, &Options{
				AllowUnknownFields: test.allowUnknown,
			}))
			if test.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t,
				yaml.Unmarshal([]byte(test.input), test.equivalent),
			)

			baseField := reflect.ValueOf(test.data).Elem().Field(0)

			// should have no unresolved value
			assert.Len(t, baseField.Interface().(BaseField).unresolvedFields, 0)

			actualData := reflect.ValueOf(test.data).Elem().Field(1).Interface()

			var expected interface{}
			if reflect.TypeOf(test.equivalent).Elem().Kind() == reflect.Struct {
				expected = reflect.ValueOf(test.equivalent).Elem().Field(0).Interface()
			} else {
				expected = reflect.ValueOf(test.equivalent).Elem().Interface()
			}

			assert.EqualValues(t, expected, actualData)
		})
	}
}

func TestBaseField_UnmarshalYAML(t *testing.T) {
	type Inline struct {
		BaseField

		StringMap map[string]string `yaml:"string_map"`
		Array     [5]interface{}    `yaml:"array"`
	}

	type Foo struct {
		BaseField

		Str     string  `yaml:"str"`
		StrPtr  *string `yaml:"str_ptr"`
		BoolPtr *bool   `yaml:"bool_ptr"`

		Other map[string]string `yaml:",inline" rs:"other"`

		InlineWithBaseField Inline `yaml:",inline"`
	}

	tests := []struct {
		name string
		yaml string

		expectedUnmarshaled *Foo
		expectedResolved    *Foo

		expectUnmarshalErr bool
	}{
		{
			name: "Basic",
			yaml: `str: bar`,

			expectedResolved: &Foo{Str: "bar"},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{unresolvedFields: nil},
				Str:       "bar",
			},
		},
		{
			name: "Basic Nil",
			yaml: `str: `,

			expectedResolved: &Foo{},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{unresolvedFields: nil},
				Str:       "",
			},
		},
		{
			name: "Basic Ptr Nil",
			yaml: `
str_ptr: null
bool_ptr: null
`,
			expectedResolved: &Foo{},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{unresolvedFields: nil},
				StrPtr:    nil,
				BoolPtr:   nil,
			},
		},
		{
			name: "Renderer",
			yaml: `str@add-suffix-test: bar`,

			expectedResolved: &Foo{Str: "bar-test"},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{
					unresolvedFields: map[string]*unresolvedFieldSpec{
						"str": {
							fieldName:   "Str",
							fieldValue:  reflect.Value{},
							rawDataList: []*alterInterface{{scalarData: "bar"}},
							renderers:   []*rendererSpec{{name: "add-suffix-test"}},
						},
					},
				},
				Str: "",
			},
		},
		{
			name: "CatchOther Duplicate Yaml key",
			yaml: `{other@echo: foo, other@echo|echo: bar}`,

			expectUnmarshalErr: true,
		},
		{
			name: "CatchOther",
			yaml: `{ other_field_1@echo: foo, other_field_2@add-suffix-test: bar }`,

			expectedResolved: &Foo{
				Other: map[string]string{
					"other_field_1": "foo",
					"other_field_2": "bar-test",
				},
			},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{
					catchOtherFields: map[string]struct{}{
						"other_field_1": {},
						"other_field_2": {},
					},
					catchOtherCache: nil,
					unresolvedFields: map[string]*unresolvedFieldSpec{
						"other_field_1": {
							fieldName:  "Other",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{
								{
									mapData: map[string]*alterInterface{
										"other_field_1": {scalarData: "foo"},
									},
								},
							},
							renderers:         []*rendererSpec{{name: "echo"}},
							isCatchOtherField: true,
						},
						"other_field_2": {
							fieldName:  "Other",
							fieldValue: reflect.Value{},
							rawDataList: []*alterInterface{{
								mapData: map[string]*alterInterface{
									"other_field_2": {scalarData: "bar"},
								},
							}},
							renderers:         []*rendererSpec{{name: "add-suffix-test"}},
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

			expectedResolved: &Foo{
				InlineWithBaseField: Inline{
					StringMap: map[string]string{"c": "e"},
					Array:     [5]interface{}{"1", "2", "3", "4", "5"},
				},
			},
			expectedUnmarshaled: &Foo{
				InlineWithBaseField: Inline{
					BaseField: BaseField{
						unresolvedFields: map[string]*unresolvedFieldSpec{
							"string_map": {
								fieldName:  "StringMap",
								fieldValue: reflect.Value{},
								rawDataList: []*alterInterface{{
									scalarData: "c: e",
								}},
								renderers: []*rendererSpec{
									{name: "echo"},
									{name: "echo"},
								},
							},
							"array": {
								fieldName:  "Array",
								fieldValue: reflect.Value{},
								rawDataList: []*alterInterface{{
									scalarData: `- "1"
- "2"
- "3"
- "4"
- '5'`,
								}},
								renderers: []*rendererSpec{
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
				out := Init(&Foo{}, nil).(*Foo)
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
				out._parentType = nil
				out.normalFields = nil
				out.catchOtherCache = nil
				out.catchOtherField = nil
				for k := range out.unresolvedFields {
					out.unresolvedFields[k].fieldValue = reflect.Value{}
				}

				// assert.EqualValues(t, 1, out.InlineWithBaseField._initialized)
				out.InlineWithBaseField._initialized = 0
				out.InlineWithBaseField._parentValue = reflect.Value{}
				out.InlineWithBaseField._parentType = nil
				out.InlineWithBaseField.normalFields = nil
				out.InlineWithBaseField.catchOtherCache = nil
				out.InlineWithBaseField.catchOtherField = nil
				for k := range out.InlineWithBaseField.unresolvedFields {
					out.InlineWithBaseField.unresolvedFields[k].fieldValue = reflect.Value{}
				}

				// expected := Init(test.expectedUnmarshaled, nil).(*Foo)
				// expected._parentValue = reflect.Value{}
				// expected.fields = nil
				// expected.catchOtherCache = nil
				// expected.InlineWithBaseField._parentValue = reflect.Value{}
				// expected.InlineWithBaseField.fields = nil
				// expected.InlineWithBaseField.catchOtherCache = nil

				assert.EqualValues(t, test.expectedUnmarshaled, out)
			})

			if test.expectUnmarshalErr {
				return
			}

			t.Run("resolve", func(t *testing.T) {
				out := Init(&Foo{}, nil).(*Foo)
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
