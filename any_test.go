package rs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// type check
var (
	_ Field          = (*AnyObject)(nil)
	_ yaml.Marshaler = (*AnyObject)(nil)
	_ json.Marshaler = (*AnyObject)(nil)

	_ Field          = (*AnyObjectMap)(nil)
	_ yaml.Marshaler = (*AnyObjectMap)(nil)
	_ json.Marshaler = (*AnyObjectMap)(nil)
)

func createExpectedJSONValue(t *testing.T, i interface{}) string {
	data, err := json.Marshal(i)
	if !assert.NoError(t, err) {
		t.FailNow()
		return ""
	}

	return string(data)
}

func TestAnyObject_NormalizedValue(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var obj *AnyObject

		t.Run("normalized", func(t *testing.T) {
			assert.Nil(t, obj.NormalizedValue())
		})

		t.Run("raw", func(t *testing.T) {
			assert.Nil(t, obj.value())
		})
	})

	t.Run("map-not-set", func(t *testing.T) {
		obj := &AnyObject{
			mapData: &AnyObjectMap{
				Data: nil,
			},
		}
		t.Run("normalized", func(t *testing.T) {
			assert.Nil(t, obj.NormalizedValue())
			assert.IsType(t, map[string]interface{}{}, obj.NormalizedValue())
		})
		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, &AnyObjectMap{}, obj.value())
		})
	})

	t.Run("map-not-nil", func(t *testing.T) {
		obj := &AnyObject{
			mapData: &AnyObjectMap{
				Data: map[string]*AnyObject{},
			},
		}
		t.Run("normalized", func(t *testing.T) {
			assert.NotNil(t, obj.NormalizedValue())
			assert.IsType(t, map[string]interface{}{}, obj.NormalizedValue())
		})
		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, &AnyObjectMap{}, obj.value())
		})
	})

	t.Run("slice-not-nil", func(t *testing.T) {
		obj := &AnyObject{
			sliceData: []*AnyObject{},
		}

		t.Run("normalized", func(t *testing.T) {
			assert.NotNil(t, obj.NormalizedValue())
			assert.IsType(t, []interface{}{}, obj.NormalizedValue())
		})
		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, []*AnyObject{}, obj.value())
		})
	})

	t.Run("original-node", func(t *testing.T) {
		obj := &AnyObject{
			originalNode: &yaml.Node{},
			scalarData:   1.1,
		}
		t.Run("normalized", func(t *testing.T) {
			assert.NotNil(t, obj.NormalizedValue())
			assert.IsType(t, 1.1, obj.NormalizedValue())
		})
		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, &yaml.Node{}, obj.value())
		})
	})

	t.Run("scalar-not-nil", func(t *testing.T) {
		obj := &AnyObject{
			scalarData: "test",
		}
		t.Run("normalized", func(t *testing.T) {
			assert.NotNil(t, obj.NormalizedValue())
			assert.IsType(t, "", obj.NormalizedValue())
		})
		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, "", obj.value())
		})
	})
}

func TestAnyObject_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string

		input    string
		expected interface{}
	}{
		{
			name:  "map",
			input: `foo: test`,
			expected: &AnyObject{
				mapData: &AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {scalarData: "test"},
					},
				},
			},
		},
		{
			name:  "seq",
			input: `[test, test]`,
			expected: &AnyObject{
				sliceData: []*AnyObject{
					{scalarData: "test"},
					{scalarData: "test"},
				},
			},
		},
		{
			name:  "str",
			input: `test`,
			expected: &AnyObject{
				scalarData: "test",
			},
		},
		{
			name:  "binary",
			input: `!!binary dGVzdA==`,
			expected: &AnyObject{
				scalarData: "dGVzdA==",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Init(&AnyObject{}, nil).(*AnyObject)
			assert.NoError(t, yaml.Unmarshal([]byte(test.input), out))
			unsetAnyObjectBaseField(out)

			assert.EqualValues(t, test.expected, out)
		})
	}
}

func TestAnyObject(t *testing.T) {
	tests := []struct {
		name  string
		input string

		expectedUnmarshaled *AnyObject
		expectedResolved    *AnyObject

		expectedEquivalent interface{}
	}{
		{
			name:  "Basic Map",
			input: `foo: bar`,

			expectedUnmarshaled: &AnyObject{
				mapData: &AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {scalarData: "bar"},
					},
				},
			},
			expectedResolved: &AnyObject{
				mapData: &AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {scalarData: "bar"},
					},
				},
			},
			expectedEquivalent: map[string]string{
				"foo": "bar",
			},
		},
		{
			name:  "Basic List",
			input: `[foo, bar]`,

			expectedUnmarshaled: &AnyObject{
				sliceData: []*AnyObject{
					{scalarData: "foo"},
					{scalarData: "bar"},
				},
			},
			expectedResolved: &AnyObject{
				sliceData: []*AnyObject{
					{scalarData: "foo"},
					{scalarData: "bar"},
				},
			},
			expectedEquivalent: []string{"foo", "bar"},
		},
		{
			name:  "Simple Rendering Suffix",
			input: `foo@echo: bar`,

			expectedUnmarshaled: &AnyObject{
				mapData: &AnyObjectMap{
					Data: nil,
				},
			},
			expectedResolved: &AnyObject{
				mapData: &AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {scalarData: "bar"},
					},
				},
			},
			expectedEquivalent: map[string]string{
				"foo": "bar",
			},
		},
		{
			name:  "Combined",
			input: `[{foo@echo: [a,b]}, {bar@echo: [c,d]}]`,

			expectedUnmarshaled: &AnyObject{
				sliceData: []*AnyObject{
					{mapData: &AnyObjectMap{Data: nil}},
					{mapData: &AnyObjectMap{Data: nil}},
				},
			},
			expectedResolved: &AnyObject{
				sliceData: []*AnyObject{
					{mapData: &AnyObjectMap{Data: map[string]*AnyObject{
						"foo": {
							sliceData: []*AnyObject{
								{scalarData: "a"},
								{scalarData: "b"},
							},
						},
					}}},
					{mapData: &AnyObjectMap{Data: map[string]*AnyObject{
						"bar": {
							sliceData: []*AnyObject{
								{scalarData: "c"},
								{scalarData: "d"},
							},
						},
					}}},
				},
			},
			expectedEquivalent: []interface{}{
				map[string]interface{}{
					"foo": []string{"a", "b"},
				},
				map[string]interface{}{
					"bar": []string{"c", "d"},
				},
			},
		},
		{
			name:  "Merging Non-nil",
			input: `foo@echo!: { value: { bar: woo }, merge: [{ value: { woo: bar } }, { value: { foo: woo } }] }`,

			expectedUnmarshaled: &AnyObject{
				mapData: &AnyObjectMap{
					Data: nil,
				},
			},
			expectedResolved: &AnyObject{
				mapData: &AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {mapData: &AnyObjectMap{
							Data: map[string]*AnyObject{
								"bar": {scalarData: "woo"},
								"woo": {scalarData: "bar"},
								"foo": {scalarData: "woo"},
							},
						}},
					},
				},
			},
			expectedEquivalent: map[string]interface{}{
				"foo": map[string]string{
					"bar": "woo",
					"woo": "bar",
					"foo": "woo",
				},
			},
		},
		{
			name:  "Merging nil",
			input: `foo@echo!: { merge: [{ value: { woo: bar } }, { value: { foo: woo } }] }`,

			expectedUnmarshaled: &AnyObject{
				mapData: &AnyObjectMap{
					Data: nil,
				},
			},
			expectedResolved: &AnyObject{
				mapData: &AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {mapData: &AnyObjectMap{
							Data: map[string]*AnyObject{
								"woo": {scalarData: "bar"},
								"foo": {scalarData: "woo"},
							},
						}},
					},
				},
			},
			expectedEquivalent: map[string]interface{}{
				"foo": map[string]string{
					"woo": "bar",
					"foo": "woo",
				},
			},
		},
	}

	rc := &testRenderingHandler{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			t.Run("unmarshal", func(t *testing.T) {
				obj := &AnyObject{}

				assert.NoError(t, yaml.Unmarshal([]byte(test.input), obj))
				unsetAnyObjectBaseField(obj)
				assert.EqualValues(t, test.expectedUnmarshaled, obj)
			})

			t.Run("resolve+marshal", func(t *testing.T) {
				obj := &AnyObject{}

				assert.NoError(t, yaml.Unmarshal([]byte(test.input), obj))
				assert.NoError(t, obj.ResolveFields(rc, -1))
				unsetAnyObjectBaseField(obj)

				assert.EqualValues(t, test.expectedResolved, obj)

				yamlResult, err := yaml.Marshal(obj)
				if !assert.NoError(t, err) {
					return
				}

				assert.Equal(t, createExpectedYamlValue(t, test.expectedEquivalent), string(yamlResult))

				jsonResult, err := json.Marshal(obj)
				if !assert.NoError(t, err) {
					return
				}

				assert.Equal(t, createExpectedJSONValue(t, test.expectedEquivalent), string(jsonResult))
			})
		})
	}
}

func TestAnyObject_WithHint(t *testing.T) {
	testAnyObjectUnmarshalAndResolveByYamlSpecs(t, "testdata/anyobject-hint")
}

func unsetAnyObjectBaseField(obj *AnyObject) {
	if obj.mapData != nil {
		obj.mapData.BaseField = BaseField{}
		for _, v := range obj.mapData.Data {
			unsetAnyObjectBaseField(v)
		}
	}

	if obj.sliceData != nil {
		for _, v := range obj.sliceData {
			unsetAnyObjectBaseField(v)
		}
	}
}
