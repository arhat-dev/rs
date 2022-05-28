package rs

import (
	"encoding/json"
	"testing"

	"arhat.dev/pkg/testhelper"
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

func createExpectedJSONValue(t *testing.T, i any) string {
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
		obj := AnyObject{
			mapData: AnyObjectMap{
				Data: nil,
			},

			kind: _mapData,
		}
		t.Run("normalized", func(t *testing.T) {
			assert.Nil(t, obj.NormalizedValue())
			assert.IsType(t, map[string]any{}, obj.NormalizedValue())
		})

		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, &AnyObjectMap{}, obj.value())
		})
	})

	t.Run("map-not-nil", func(t *testing.T) {
		obj := AnyObject{
			mapData: AnyObjectMap{
				Data: map[string]*AnyObject{},
			},

			kind: _mapData,
		}
		t.Run("normalized", func(t *testing.T) {
			assert.NotNil(t, obj.NormalizedValue())
			assert.IsType(t, map[string]any{}, obj.NormalizedValue())
		})
		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, &AnyObjectMap{}, obj.value())
		})
	})

	t.Run("slice-not-nil", func(t *testing.T) {
		obj := AnyObject{
			sliceData: []AnyObject{},

			kind: _sliceData,
		}

		t.Run("normalized", func(t *testing.T) {
			assert.NotNil(t, obj.NormalizedValue())
			assert.IsType(t, []any{}, obj.NormalizedValue())
		})
		t.Run("raw", func(t *testing.T) {
			assert.NotNil(t, obj.value())
			assert.IsType(t, []*AnyObject{}, obj.value())
		})
	})

	t.Run("scalar-not-nil", func(t *testing.T) {
		obj := AnyObject{
			scalarData: "test",

			kind: _scalarData,
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
		expected any
	}{
		{
			name:  "map",
			input: `foo: test`,
			expected: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {
							scalarData: "test",
							kind:       _scalarData,
						},
					},
				},
			},
		},
		{
			name:  "seq",
			input: `[test, test]`,
			expected: &AnyObject{
				kind: _sliceData,
				sliceData: []AnyObject{
					{scalarData: "test", kind: _scalarData},
					{scalarData: "test", kind: _scalarData},
				},
			},
		},
		{
			name:  "str",
			input: `test`,
			expected: &AnyObject{
				scalarData: "test",
				kind:       _scalarData,
			},
		},
		{
			name:  "binary",
			input: `!!binary dGVzdA==`,
			expected: &AnyObject{
				scalarData: "dGVzdA==",
				kind:       _scalarData,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Init(&AnyObject{}, nil)
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

		expectedEquivalent any
	}{
		{
			name:  "Basic Map",
			input: `foo: bar`,

			expectedUnmarshaled: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {scalarData: "bar", kind: _scalarData},
					},
				},
			},
			expectedResolved: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {scalarData: "bar", kind: _scalarData},
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
				kind: _sliceData,
				sliceData: []AnyObject{
					{scalarData: "foo", kind: _scalarData},
					{scalarData: "bar", kind: _scalarData},
				},
			},
			expectedResolved: &AnyObject{
				kind: _sliceData,
				sliceData: []AnyObject{
					{scalarData: "foo", kind: _scalarData},
					{scalarData: "bar", kind: _scalarData},
				},
			},
			expectedEquivalent: []string{"foo", "bar"},
		},
		{
			name:  "Simple Rendering Suffix",
			input: `foo@echo: bar`,

			expectedUnmarshaled: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: nil,
				},
			},
			expectedResolved: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {scalarData: "bar", kind: _scalarData},
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
				kind: _sliceData,
				sliceData: []AnyObject{
					{mapData: AnyObjectMap{Data: nil}, kind: _mapData},
					{mapData: AnyObjectMap{Data: nil}, kind: _mapData},
				},
			},
			expectedResolved: &AnyObject{
				kind: _sliceData,
				sliceData: []AnyObject{
					{
						kind: _mapData,
						mapData: AnyObjectMap{
							Data: map[string]*AnyObject{
								"foo": {
									kind: _sliceData,
									sliceData: []AnyObject{
										{scalarData: "a", kind: _scalarData},
										{scalarData: "b", kind: _scalarData},
									},
								},
							},
						},
					},
					{
						kind: _mapData,
						mapData: AnyObjectMap{
							Data: map[string]*AnyObject{
								"bar": {
									kind: _sliceData,
									sliceData: []AnyObject{
										{scalarData: "c", kind: _scalarData},
										{scalarData: "d", kind: _scalarData},
									},
								},
							},
						},
					},
				},
			},
			expectedEquivalent: []any{
				map[string]any{
					"foo": []string{"a", "b"},
				},
				map[string]any{
					"bar": []string{"c", "d"},
				},
			},
		},
		{
			name:  "Merge Non-nil",
			input: `foo@echo!: { value: { bar: woo }, merge: [{ value: { woo: bar } }, { value: { foo: woo } }] }`,

			expectedUnmarshaled: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: nil,
				},
			},
			expectedResolved: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {
							kind: _mapData,
							mapData: AnyObjectMap{
								Data: map[string]*AnyObject{
									"bar": {scalarData: "woo", kind: _scalarData},
									"woo": {scalarData: "bar", kind: _scalarData},
									"foo": {scalarData: "woo", kind: _scalarData},
								},
							},
						},
					},
				},
			},
			expectedEquivalent: map[string]any{
				"foo": map[string]string{
					"bar": "woo",
					"woo": "bar",
					"foo": "woo",
				},
			},
		},
		{
			name:  "Merge nil",
			input: `foo@echo!: { merge: [{ value: { woo: bar } }, { value: { foo: woo } }] }`,

			expectedUnmarshaled: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: nil,
				},
			},
			expectedResolved: &AnyObject{
				kind: _mapData,
				mapData: AnyObjectMap{
					Data: map[string]*AnyObject{
						"foo": {
							kind: _mapData,
							mapData: AnyObjectMap{
								Data: map[string]*AnyObject{
									"woo": {scalarData: "bar", kind: _scalarData},
									"foo": {scalarData: "woo", kind: _scalarData},
								},
							},
						},
					},
				},
			},
			expectedEquivalent: map[string]any{
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

			t.Run("Unmarshal", func(t *testing.T) {
				obj := &AnyObject{}

				assert.NoError(t, yaml.Unmarshal([]byte(test.input), obj))
				unsetAnyObjectBaseField(obj)
				assert.EqualValues(t, test.expectedUnmarshaled, obj)
			})

			t.Run("Resolve and Marshal", func(t *testing.T) {
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
	if obj == nil {
		return
	}

	obj.BaseField = BaseField{}
	switch obj.kind {
	case _mapData:
		obj.mapData.BaseField = BaseField{}
		for _, v := range obj.mapData.Data {
			unsetAnyObjectBaseField(v)
		}
	case _sliceData:
		for i := range obj.sliceData {
			unsetAnyObjectBaseField(&obj.sliceData[i])
		}
	}
}

func testAnyObjectUnmarshalAndResolveByYamlSpecs(t *testing.T, specDirPath string) {
	testhelper.TestFixtures(t, specDirPath,
		func() any { return InitAny(&AnyObject{}, nil) },
		func() any { var out any; return &out },
		func(t *testing.T, spec, exp any) {
			testSrc := spec.(*AnyObject)
			exp = *exp.(*any)

			assert.NoError(t,
				testSrc.ResolveFields(testRenderingHandler{}, -1),
				"failed to resolve test source",
			)

			ret, err := yaml.Marshal(testSrc)
			assert.NoError(t, err, "failed to marshal resolved data")

			var actual any
			assert.NoError(t, yaml.Unmarshal(ret, &actual), "failed to unmarshal resolved data")

			assert.EqualValues(t, exp, actual)
		},
	)
}
