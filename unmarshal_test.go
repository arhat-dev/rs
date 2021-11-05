package rs

import (
	"encoding/base64"
	"fmt"
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
			assert.Len(t, baseField.Interface().(BaseField).unresolvedNormalFields, 0)

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

	type InlineDelegated struct {
		StringMap map[string]string `yaml:"delegated_string_map"`
		Array     [5]interface{}    `yaml:"delegated_array"`
	}

	type Foo struct {
		BaseField

		Str     string  `yaml:"str"`
		StrPtr  *string `yaml:"str_ptr"`
		BoolPtr *bool   `yaml:"bool_ptr"`

		Other map[string]string `yaml:",inline" rs:"other"`

		InlineWithBaseField Inline `yaml:",inline"`

		InlineWithoutBaseField InlineDelegated `yaml:",inline"`
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
				BaseField: BaseField{
					unresolvedNormalFields: nil,
				},
				Str: "bar",
			},
		},
		{
			name: "Basic Nil",
			yaml: `str: `,

			expectedResolved: &Foo{},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{
					unresolvedNormalFields: nil,
				},
				Str: "",
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
				BaseField: BaseField{
					unresolvedNormalFields: nil,
				},
				StrPtr:  nil,
				BoolPtr: nil,
			},
		},
		{
			name: "Renderer",
			yaml: `str@add-suffix-test: bar`,

			expectedResolved: &Foo{Str: "bar-test"},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{
					unresolvedNormalFields: map[string]*unresolvedFieldSpec{
						"str": {
							fieldName:  "Str",
							fieldValue: reflect.Value{},
							rawData:    fakeScalarNode("bar"),
							renderers:  []*rendererSpec{{name: "add-suffix-test"}},
						},
					},
				},
				Str: "",
			},
		},
		{
			name: "CatchOther Can Have Duplicate Yaml key",
			yaml: `{other@echo: foo, other@echo: bar}`,

			expectedResolved: &Foo{
				Other: map[string]string{
					"other": "bar",
				},
			},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{
					unresolvedInlineMapItems: map[string][]*unresolvedFieldSpec{
						"other": {
							{
								fieldName:       "other",
								fieldValue:      reflect.Value{},
								rawData:         fakeMap(fakeScalarNode("other"), fakeScalarNode("foo")),
								renderers:       []*rendererSpec{{name: "echo"}},
								isInlineMapItem: true,
							},
							{
								fieldName:       "other",
								fieldValue:      reflect.Value{},
								rawData:         fakeMap(fakeScalarNode("other"), fakeScalarNode("bar")),
								renderers:       []*rendererSpec{{name: "echo"}},
								isInlineMapItem: true,
							},
						},
					},
				},
			},
		},
		{
			name: "CatchOther",
			yaml: `{ other_field_1@echo: foo, other_field_2@add-suffix-test: bar }`,

			expectedResolved: &Foo{
				BaseField: BaseField{},
				Other: map[string]string{
					"other_field_1": "foo",
					"other_field_2": "bar-test",
				},
			},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{
					// unresolvedInlineMapKeys: map[string]struct{}{
					// 	"other_field_1": {},
					// 	"other_field_2": {},
					// },
					// unresolvedNormalFields: map[string]*unresolvedFieldSpec{},
					unresolvedInlineMapItems: map[string][]*unresolvedFieldSpec{
						"other_field_1": {{
							fieldName:       "other_field_1",
							fieldValue:      reflect.Value{},
							rawData:         fakeMap(fakeScalarNode("other_field_1"), fakeScalarNode("foo")),
							renderers:       []*rendererSpec{{name: "echo"}},
							isInlineMapItem: true,
						}},
						"other_field_2": {{
							fieldName:       "other_field_2",
							fieldValue:      reflect.Value{},
							rawData:         fakeMap(fakeScalarNode("other_field_2"), fakeScalarNode("bar")),
							renderers:       []*rendererSpec{{name: "add-suffix-test"}},
							isInlineMapItem: true,
						}},
					},
				},
				// `Other` field should NOT be initialized
				// it will be initialized during resolving
				Other: nil,
			},
		},
		{
			name: "Inline",
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
				BaseField: BaseField{},
				InlineWithBaseField: Inline{
					StringMap: map[string]string{"c": "e"},
					Array:     [5]interface{}{"1", "2", "3", "4", "5"},
				},
			},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{},
				InlineWithBaseField: Inline{
					BaseField: BaseField{
						unresolvedNormalFields: map[string]*unresolvedFieldSpec{
							"string_map": {
								fieldName:  "StringMap",
								fieldValue: reflect.Value{},
								rawData:    fakeScalarNode("c: e"),
								renderers: []*rendererSpec{
									{name: "echo"},
									{name: "echo"},
								},
							},
							"array": {
								fieldName:  "Array",
								fieldValue: reflect.Value{},
								rawData: fakeScalarNode(`- "1"
- "2"
- "3"
- "4"
- '5'`),
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
		{
			name: "Inline Delegated",
			// editorconfig-checker-disable
			yaml: `---

delegated_string_map@echo|echo: |-
  c: e
delegated_array@echo|echo|echo: |-
  - "1"
  - "2"
  - "3"
  - "4"
  - '5'
`,
			// editorconfig-checker-enable

			expectedResolved: &Foo{
				BaseField: BaseField{},
				InlineWithoutBaseField: InlineDelegated{
					StringMap: map[string]string{"c": "e"},
					Array:     [5]interface{}{"1", "2", "3", "4", "5"},
				},
			},
			expectedUnmarshaled: &Foo{
				BaseField: BaseField{
					unresolvedNormalFields: map[string]*unresolvedFieldSpec{
						"delegated_string_map": {
							fieldName:  "StringMap",
							fieldValue: reflect.Value{},
							rawData:    fakeScalarNode("c: e"),
							renderers: []*rendererSpec{
								{name: "echo"},
								{name: "echo"},
							},
						},
						"delegated_array": {
							fieldName:  "Array",
							fieldValue: reflect.Value{},
							rawData: fakeScalarNode(`- "1"
- "2"
- "3"
- "4"
- '5'`),
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
				cleanupBaseField(t, &out.BaseField)
				cleanupBaseField(t, &out.InlineWithBaseField.BaseField)

				// expected := Init(test.expectedUnmarshaled, nil).(*Foo)
				// expected._parentValue = reflect.Value{}
				// expected.fields = nil
				// expected.inlineMapCache = nil
				// expected.InlineWithBaseField._parentValue = reflect.Value{}
				// expected.InlineWithBaseField.fields = nil
				// expected.InlineWithBaseField.inlineMapCache = nil

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

// ensure we maintain the same behavior when unmarshal multiple times
func TestBaseField_UnmarshalYAML_multiple_times(t *testing.T) {
	// editorconfig-checker-disable
	const InitialInput = `
int: 1
int_ptr: 2

slice: [foo]
ptr_slice: [foo]

array: [foo, bar]
ptr_array: [foo, bar]

struct: {foo: foo}
struct_ptr: {foo: foo}

any_map:
  foo: bar
ptr_map:
  foo: bar
`

	const SecondInput = `
int:
int_ptr:

slice:
ptr_slice:

array:
ptr_array:

struct:
struct_ptr:

any_map:
  foo:

ptr_map:
  foo:
`

	const ThirdInput = `
int: 3
int_ptr: 4

slice: []
ptr_slice: []

array: [a, b]
ptr_array: [a, b]

struct: {}
struct_ptr: {}

any_map:
ptr_map:
`
	// editorconfig-checker-enable
	type TestCase struct {
		Int    int  `yaml:"int"`
		IntPtr *int `yaml:"int_ptr"`

		Slice    []string  `yaml:"slice"`
		PtrSlice []*string `yaml:"ptr_slice"`

		Array    [2]string  `yaml:"array"`
		PtrArray [2]*string `yaml:"ptr_array"`

		Struct struct {
			Foo string `yaml:"foo"`
		} `yaml:"struct"`

		StructPtr *struct {
			Foo string `yaml:"foo"`
		} `yaml:"struct_ptr"`

		AnyMap map[string]interface{} `yaml:"any_map"`
		PtrMap map[string]*string     `yaml:"ptr_map"`
	}

	var (
		outRS struct {
			BaseField
			TestCase `yaml:",inline"`
		}
		outExpected TestCase
	)

	_ = Init(&outRS, nil)

	assertVisibleValues := func(t *testing.T, expected, actual *TestCase) {
		assert.EqualValues(t, expected.Int, actual.Int)
		assert.EqualValues(t, expected.IntPtr, actual.IntPtr)

		assert.EqualValues(t, expected.Slice, actual.Slice)
		assert.EqualValues(t, expected.PtrSlice, actual.PtrSlice)

		assert.EqualValues(t, expected.Array, actual.Array)
		assert.EqualValues(t, expected.PtrArray, actual.PtrArray)

		assert.EqualValues(t, expected.Struct, actual.Struct)
		assert.EqualValues(t, expected.StructPtr, actual.StructPtr)

		assert.EqualValues(t, expected.AnyMap, actual.AnyMap)
		assert.EqualValues(t, expected.PtrMap, actual.PtrMap)
	}

	for i, test := range []string{InitialInput, SecondInput, ThirdInput} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert.NoError(t, yaml.Unmarshal([]byte(test), &outRS))
			assert.NoError(t, yaml.Unmarshal([]byte(test), &outExpected))

			t.Log("expected =", outExpected, "actual =", outRS.TestCase)
			assertVisibleValues(t, &outExpected, &outRS.TestCase)
		})
	}
}

func fakeScalarNode(i interface{}) *yaml.Node {
	var (
		data []byte
		tag  string
	)
	switch v := i.(type) {
	case bool:
		tag = boolTag
	case string:
		tag = strTag
		data = []byte(v)
	case []byte:
		tag = binaryTag
		data = []byte(tag + " " + base64.StdEncoding.EncodeToString(v))
	case float32, float64:
		tag = floatTag
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr:
		tag = intTag
	default:
		panic(fmt.Errorf("invalid non scalar value %T: %v", i, i))
	}

	switch vt := i.(type) {
	case string, []byte:
		// already set before
	default:
		var err error
		data, err = yaml.Marshal(vt)
		if err != nil {
			panic(err)
		}
	}

	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   tag,
		Value: string(data),
	}
}

func cleanupBaseField(t *testing.T, f *BaseField) {
	_ = t
	if f == nil {
		return
	}

	f._initialized = 0
	f._parentValue = reflect.Value{}
	f._parentType = nil
	f.normalFields = nil
	f.inlineMap = nil
	for k := range f.unresolvedNormalFields {
		f.unresolvedNormalFields[k].fieldValue = reflect.Value{}
		cleanupYamlNode(f.unresolvedNormalFields[k].rawData)
	}
	for _, list := range f.unresolvedInlineMapItems {
		for _, v := range list {
			v.fieldValue = reflect.Value{}
			cleanupYamlNode(v.rawData)
		}
	}
}

func cleanupYamlNode(n *yaml.Node) {
	for _, cn := range n.Content {
		cleanupYamlNode(cn)
	}

	n.Line = 0
	n.Style = 0
	n.Column = 0
}

func TestBaseField_UnmarshalYAML_Mixed_CatchOther(t *testing.T) {
	type Foo struct {
		BaseField

		Data map[string][]string `rs:"other"`
	}

	a := Init(&Foo{}, nil).(*Foo)
	assert.NoError(t, yaml.Unmarshal([]byte(`{ a: [a], b@echo: [b], c: [c], b@echo: [x] }`), a))
	assert.EqualValues(t, map[string][]string{
		"a": {"a"},
		"c": {"c"},
	}, a.Data)

	assert.NoError(t, a.ResolveFields(&testRenderingHandler{}, -1))
	assert.EqualValues(t, map[string][]string{
		"a": {"a"},
		"b": {"b", "x"},
		"c": {"c"},
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
		assert.Len(t, out.BaseField.unresolvedNormalFields, 1)
		assert.Len(t, out.Foo.BaseField.unresolvedNormalFields, 0)
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
		assert.Len(t, out.BaseField.unresolvedNormalFields, 0)
		assert.Len(t, out.Foo.BaseField.unresolvedNormalFields, 1)
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

func TestRS_Tag_Disabled(t *testing.T) {
	type TestCase struct {
		BaseField

		Enabled  interface{} `yaml:"enabled" rs:""`
		Disabled interface{} `yaml:"disabled" rs:"disabled"`
	}

	out := Init(&TestCase{}, nil).(*TestCase)
	t.Run("Good", func(t *testing.T) {
		assert.NoError(t, yaml.Unmarshal([]byte(`{ enabled@foo: foo, disabled: ok }`), out))
	})

	t.Run("Bad", func(t *testing.T) {
		assert.Error(t, yaml.Unmarshal([]byte(`{ enabled: foo, disabled@foo: bar }`), out))
	})

	t.Run("Invalid", func(t *testing.T) {
		assert.Error(t, yaml.Unmarshal([]byte(`{ enabled: foo, some@foo: bar }`), out))
	})
}
