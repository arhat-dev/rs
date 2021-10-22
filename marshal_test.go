package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// type check
var (
	_ yaml.Marshaler = (*BaseField)(nil)
)

type marshalTestSpec struct {
	name string

	data interface{}

	inputNoRS   string
	inputWithRS string

	equivalent interface{}
}

var (
	_ Field          = (*dataWrapper)(nil)
	_ yaml.Marshaler = (*dataWrapper)(nil)
)

type dataWrapper struct {
	data interface{}
}

func (dw *dataWrapper) ResolveFields(rc RenderingHandler, depth int, fieldNames ...string) error {
	bf := reflect.ValueOf(dw.data).Elem().Field(0).Interface().(BaseField)
	return (*BaseField).ResolveFields(&bf, rc, depth, fieldNames...)
}

func (dw *dataWrapper) UnmarshalYAML(n *yaml.Node) error {
	bf := reflect.ValueOf(dw.data).Elem().Field(0).Interface().(BaseField)
	return (*BaseField).UnmarshalYAML(&bf, n)
}

func (dw *dataWrapper) MarshalYAML() (interface{}, error) {
	bf := reflect.ValueOf(dw.data).Elem().Field(0).Interface().(BaseField)
	return (*BaseField).MarshalYAML(&bf)
}

func runMarshalTest(t *testing.T, tests []marshalTestSpec) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedBytes, err := yaml.Marshal(test.equivalent)
			assert.NoError(t, err)

			expected := string(expectedBytes)

			t.Run("Direct Set", func(t *testing.T) {
				ret, err := yaml.Marshal(initInterface(test.data, nil))
				assert.NoError(t, err)

				t.Log(string(ret))
				assert.EqualValues(t, expected, string(ret))
			})

			newEmptyValue := func() Field {
				return &dataWrapper{
					data: initInterface(
						reflect.New(reflect.TypeOf(test.data).Elem()).Interface(),
						nil,
					),
				}
			}

			t.Run("After Unmarshal", func(t *testing.T) {
				v := Init(newEmptyValue(), nil)
				assert.NoError(t, yaml.Unmarshal([]byte(test.inputNoRS), v))

				ret, err := yaml.Marshal(v)
				assert.NoError(t, err)

				assert.EqualValues(t, expected, string(ret))
			})

			t.Run("After Unmarshal And Resolve", func(t *testing.T) {
				v := Init(newEmptyValue(), nil)
				assert.NoError(t, yaml.Unmarshal([]byte(test.inputWithRS), v))

				assert.NoError(t, v.ResolveFields(&testRenderingHandler{}, -1))

				ret, err := yaml.Marshal(v)
				assert.NoError(t, err)

				assert.EqualValues(t, expected, string(ret))
			})
		})
	}
}

func TestBaseField_MarshalYAML_uninitialized(t *testing.T) {
	_, err := yaml.Marshal(&struct {
		BaseField
		Foo string
	}{Foo: ""})

	assert.Error(t, err)
}

func TestBaseField_MarshalYAML_primitive(t *testing.T) {
	var allTests []marshalTestSpec

	type minorSpec struct {
		nameSuffix string
		actualType reflect.Type
		tag        string

		value reflect.Value
	}

	basicMinorSpecs := func(majorDataType reflect.Type, testVal reflect.Value, testValPtr reflect.Value) []minorSpec {
		ret := []minorSpec{
			{
				nameSuffix: "Not Empty",
				actualType: majorDataType,
				tag:        ``,

				value: testVal,
			},
			{
				nameSuffix: "Not Empty Omitempty",
				actualType: majorDataType,
				tag:        ``,

				value: testVal,
			},
			{
				nameSuffix: "Empty",
				actualType: majorDataType,
				tag:        ``,

				value: reflect.Zero(majorDataType),
			},
			{
				nameSuffix: "Empty Omitempty",
				actualType: majorDataType,
				tag:        `yaml:",omitempty"`,

				value: reflect.Zero(majorDataType),
			},
			{
				nameSuffix: "Ptr Nil",
				actualType: reflect.PtrTo(majorDataType),
				tag:        ``,

				value: reflect.Zero(reflect.PtrTo(majorDataType)),
			},
			{
				nameSuffix: "Ptr Nil Omitempty",
				actualType: reflect.PtrTo(majorDataType),
				tag:        `yaml:",omitempty"`,

				value: reflect.Zero(reflect.PtrTo(majorDataType)),
			},

			{
				nameSuffix: "Ptr Not Nil",
				actualType: reflect.PtrTo(majorDataType),
				tag:        ``,

				value: testValPtr,
			},
			{
				nameSuffix: "Ptr Not Nil Omitempty",
				actualType: reflect.PtrTo(majorDataType),
				tag:        `yaml:",omitempty"`,

				value: testValPtr,
			},
		}

		switch majorDataType.Kind() {
		case reflect.Slice, reflect.Interface, reflect.Map:
			ret = append([]minorSpec{
				{
					nameSuffix: "Nil",
					actualType: majorDataType,
					value:      reflect.ValueOf(nil),
				},
				{
					nameSuffix: "Nil Omitempty",
					actualType: majorDataType,
					tag:        `yaml:",omitempty"`,
					value:      reflect.ValueOf(nil),
				},
			}, ret...)
		}

		return ret
	}

	for _, testMajor := range []interface{}{
		string("str-value"),
		bool(true),
		float32(1.1), float64(1.1),
		int(1), int8(1), int16(1), int32(1), rune('D'), int64(1),
		uint(1), byte(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),

		// go-yaml doesn't support marshaling of complex values
		//
		// complex64(complex(float32(1.1), float32(1.1))),
		// complex128(complex(float64(1.1), float64(1.1))),

		[]interface{}{
			string("str-value"),
			bool(true),
			float32(1.1), float64(1.1),
			int(1), int8(1), int16(1), int32(1), rune('D'), int64(1),
			uint(1), byte(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),
		},
		[...]interface{}{
			string("str-value"),
			bool(true),
			float32(1.1), float64(1.1),
			int(1), int8(1), int16(1), int32(1), rune('D'), int64(1),
			uint(1), byte(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),
		},
		map[string]interface{}{
			"string":  string("str-value"),
			"bool":    bool(true),
			"float32": float32(1.1),
			"float64": float64(1.1),
			"int":     int(1),
			"int8":    int8(1),
			"int16":   int16(1),
			"int32":   int32(1),
			"rune":    rune('D'),
			"int64":   int64(1),
			"uint":    uint(1),
			"byte":    byte(1),
			"uint8":   uint8(1),
			"uint16":  uint16(1),
			"uint32":  uint32(1),
			"uint64":  uint64(1),
			"uintptr": uintptr(1),
		},
		map[string][]interface{}{
			"foo": {
				string("str-value"),
				bool(true),
				float32(1.1), float64(1.1),
				int(1), int8(1), int16(1), int32(1), rune('D'), int64(1),
				uint(1), byte(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),
			},
		},
		map[string]map[string][]interface{}{
			"foo": {"foo": {
				string("str-value"),
				bool(true),
				float32(1.1), float64(1.1),
				int(1), int8(1), int16(1), int32(1), rune('D'), int64(1),
				uint(1), byte(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),
			}},
			"bar": {"bar": {
				string("str-value"),
				bool(true),
				float32(1.1), float64(1.1),
				int(1), int8(1), int16(1), int32(1), rune('D'), int64(1),
				uint(1), byte(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),
			}},
		},
	} {

		majorDataType := reflect.TypeOf(testMajor)

		testVal := reflect.ValueOf(testMajor)
		testValPtr := reflect.New(majorDataType)
		testValPtr.Elem().Set(testVal)

		minorSpecs := basicMinorSpecs(majorDataType, testVal, testValPtr)
		for _, test := range minorSpecs {
			typePlain := reflect.StructOf([]reflect.StructField{
				{
					Name:      "Data",
					Type:      test.actualType,
					Tag:       reflect.StructTag(test.tag),
					Anonymous: false,
				},
			})

			typeWithBaseField := reflect.StructOf([]reflect.StructField{
				{
					Name:      "BaseField",
					Type:      baseFieldStructType,
					Tag:       `yaml:"-"`,
					Anonymous: true,
				},
				{
					Name:      "Data",
					Type:      test.actualType,
					Tag:       reflect.StructTag(test.tag),
					Anonymous: false,
				},
			})

			data := reflect.New(typeWithBaseField)
			equivalent := reflect.New(typePlain)
			var valueData interface{}
			if test.value.IsValid() {
				valueData = test.value.Interface()
				data.Elem().Field(1).Set(test.value)
				equivalent.Elem().Field(0).Set(test.value)
			} else {
				valueData = nil
			}

			valueBytes, err := yaml.Marshal(map[string]interface{}{
				"data": valueData,
			})
			assert.NoError(t, err)

			rsValueBytes, err := yaml.Marshal(map[string]interface{}{
				"data@echo": valueData,
			})
			assert.NoError(t, err)

			allTests = append(allTests, marshalTestSpec{
				name: majorDataType.String() + " " + test.nameSuffix,
				data: data.Interface(),

				inputNoRS:   string(valueBytes),
				inputWithRS: string(rsValueBytes),

				equivalent: equivalent.Interface(),
			})
		}
	}

	runMarshalTest(t, allTests)
}

func TestBaseField_MarshalYAML(t *testing.T) {

	type (
		Foo              struct{ Foo string }
		FooWithBaseField struct {
			BaseField
			Foo string
		}

		_l3_embedded struct {
			L3Value string
		}
		_l2_embedded struct {
			L2Value _l3_embedded `yaml:",inline"`
		}
		MultiLevelEmbedded struct {
			L1Value _l2_embedded `yaml:",inline"`
		}

		_l3_embedded_rs struct {
			BaseField
			L3Value string
		}
		_l2_embedded_rs struct {
			BaseField
			L2Value _l3_embedded_rs `yaml:",inline"`
		}
		MultiLevelEmbeddedWithBaseField struct {
			BaseField
			L1Value _l2_embedded_rs `yaml:",inline"`
		}
	)

	tests := []marshalTestSpec{
		{
			// to address https://github.com/go-yaml/yaml/issues/362
			name: "Inline Struct",
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

			inputNoRS:   "foo: foo",
			inputWithRS: "foo@echo: foo",

			equivalent: &struct {
				Struct struct {
					Foo string
				} `yaml:",inline"`
			}{Struct: struct{ Foo string }{Foo: "foo"}},
		},
		{
			// to address https://github.com/go-yaml/yaml/issues/362
			name: "Embedded Inline Struct",
			data: &struct {
				BaseField
				FooWithBaseField `yaml:",inline"`
			}{FooWithBaseField: FooWithBaseField{Foo: "foo"}},

			inputNoRS:   "foo: foo",
			inputWithRS: "foo@echo: foo",

			equivalent: &struct {
				Foo `yaml:",inline"`
			}{Foo: Foo{Foo: "foo"}},
		},

		{
			name: "Multi Level Embedded Inline Struct",
			data: &MultiLevelEmbeddedWithBaseField{
				L1Value: _l2_embedded_rs{
					L2Value: _l3_embedded_rs{
						L3Value: "foo",
					},
				},
			},

			inputNoRS:   "l3value: foo",
			inputWithRS: "l3value@echo: foo",

			equivalent: &MultiLevelEmbedded{
				L1Value: _l2_embedded{
					L2Value: _l3_embedded{
						L3Value: "foo",
					},
				},
			},
		},

		{
			name: "Interface Nil",
			data: &struct {
				BaseField
				IFace yaml.Marshaler
			}{IFace: nil},

			inputNoRS:   "iface: null",
			inputWithRS: "iface@echo: ",

			equivalent: &struct {
				IFace yaml.Marshaler
			}{IFace: nil},
		},
		{
			name: "Struct Ptr Nil Panic if not check value kind",
			data: &struct {
				BaseField
				Ptr *FooWithBaseField
			}{Ptr: nil},

			inputNoRS:   "ptr: ",
			inputWithRS: "ptr@echo: null",

			equivalent: &struct {
				Ptr *Foo
			}{Ptr: nil},
		},
		{
			name: "Struct Ptr Nil Not Panic When omitempty",
			data: &struct {
				BaseField
				Ptr *FooWithBaseField `yaml:",omitempty"`
			}{Ptr: nil},

			inputNoRS:   "ptr: null",
			inputWithRS: "ptr@echo: ",

			equivalent: &struct {
				Ptr *Foo `yaml:",omitempty"`
			}{Ptr: nil},
		},
		{
			name: "Catch Other Only",
			data: &struct {
				BaseField
				Data map[string]string `rs:"other"`
			}{Data: map[string]string{"a": "b", "c": "d"}},

			inputNoRS:   "{a: b, c: d}",
			inputWithRS: "{a@echo: b, c@echo: d}",

			equivalent: map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "Catch Other",
			data: &struct {
				BaseField
				Bar string `yaml:"bar"`

				Data map[string]string `rs:"other"`
			}{Bar: "foo", Data: map[string]string{"a": "b", "c": "d"}},

			inputNoRS:   "{ bar: foo, a: b, c: d}",
			inputWithRS: "{ bar@echo: foo, a@echo: b, c@echo: d}",

			equivalent: map[string]string{
				"bar": "foo",
				"a":   "b",
				"c":   "d",
			},
		},
		{
			name: "Catch Other Not Nil",
			data: &struct {
				BaseField

				Data map[string]*FooWithBaseField `rs:"other"`
			}{Data: map[string]*FooWithBaseField{
				"a": Init(&FooWithBaseField{Foo: "b"}, nil).(*FooWithBaseField),
				"c": Init(&FooWithBaseField{Foo: "d"}, nil).(*FooWithBaseField),
			}},

			inputNoRS:   "{ a: {foo: b}, c: {foo: d}}",
			inputWithRS: "{ a@echo: {foo@echo: b}, c@echo: {foo@echo: d}}",

			equivalent: map[string]*Foo{
				"a": {"b"},
				"c": {"d"},
			},
		},
		{
			name: "Catch Other Nil",
			data: &struct {
				BaseField

				Data map[string]*FooWithBaseField `rs:"other"`
			}{Data: map[string]*FooWithBaseField{
				"a": nil,
				"c": nil,
			}},

			inputNoRS:   "{ a: null, c: }",
			inputWithRS: "{ a@echo: null, c@echo: }",

			equivalent: map[string]*Foo{
				"a": nil,
				"c": nil,
			},
		},
	}

	runMarshalTest(t, tests)
}
