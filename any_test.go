package rs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func createExpectedJSONValue(t *testing.T, i interface{}) string {
	data, err := json.Marshal(i)
	if !assert.NoError(t, err) {
		t.FailNow()
		return ""
	}

	return string(data)
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
