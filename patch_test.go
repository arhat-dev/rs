package rs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMergeMap(t *testing.T) {
	tests := []struct {
		name string

		original   map[string]interface{}
		additional map[string]interface{}
		unique     bool

		expectErr bool
		expected  map[string]interface{}
	}{
		{
			name:       "Simple Nop",
			original:   map[string]interface{}{"foo": "bar"},
			additional: nil,
			expected:   map[string]interface{}{"foo": "bar"},
		},
		{
			name:       "Simple Merge",
			original:   map[string]interface{}{"foo": "bar"},
			additional: map[string]interface{}{"foo": "bar"},
			expected:   map[string]interface{}{"foo": "bar"},
		},
		// TODO: Add complex test cases
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := MergeMap(
				test.original, test.additional, true, test.unique,
			)
			if test.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, result)
		})
	}
}

func TestUniqueList(t *testing.T) {
	mapVal := map[string]interface{}{
		"foo": "bar",
		"bar": map[string]interface{}{
			"foo": "bar",
		},
	}
	tests := []struct {
		name string

		input    []interface{}
		expected []interface{}
	}{
		{
			name:     "Simple String",
			input:    []interface{}{"a", "c", "c", "a"},
			expected: []interface{}{"a", "c"},
		},
		{
			name:     "Simple Number",
			input:    []interface{}{1, 1, 1, 1, 1},
			expected: []interface{}{1},
		},
		{
			name:     "Map Value",
			input:    []interface{}{mapVal, mapVal, 1},
			expected: []interface{}{mapVal, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.EqualValues(t, test.expected, UniqueList(test.input))
		})
	}
}

func createMergeValue(t *testing.T, i interface{}) []MergeSource {
	data, err := yaml.Marshal(i)
	if !assert.NoError(t, err) {
		t.FailNow()
		return nil
	}

	ret := new(AnyObject)
	if !assert.NoError(t, yaml.Unmarshal(data, &ret)) {
		t.FailNow()
		return nil
	}

	return []MergeSource{{Value: ret}}
}

func createExpectedYamlValue(t *testing.T, i interface{}) string {
	data, err := yaml.Marshal(i)
	if !assert.NoError(t, err) {
		t.FailNow()
		return ""
	}

	return string(data)
}

func TestPatchSpec_ApplyTo(t *testing.T) {
	tests := []struct {
		name string

		spec     renderingPatchSpec
		original string

		expectErr bool
		expected  interface{}
	}{
		{
			name:     "Valid Nop List Merge",
			spec:     renderingPatchSpec{},
			original: `[a, b, c]`,
			expected: []string{"a", "b", "c"},
		},
		{
			name: "Valid List Merge Only",
			spec: renderingPatchSpec{
				Merge: createMergeValue(t, []string{"a", "b", "c"}),
			},
			original: ``,
			expected: []string{"a", "b", "c"},
		},
		{
			name: "Invalid List Merge Type Not Match",
			spec: renderingPatchSpec{
				Merge: createMergeValue(t, "oops: not a list"),
			},
			original:  `[a, b, c]`,
			expectErr: true,
		},
		{
			name: "List Merge",
			spec: renderingPatchSpec{
				Merge: createMergeValue(t, []string{"c", "d", "e", "f"}),
			},
			original: `[a, b, c]`,
			expected: []string{
				"a", "b", "c",
				"c", // expected dup
				"d", "e", "f",
			},
		},
		{
			name: "List Merge Unique",
			spec: renderingPatchSpec{
				Merge:  createMergeValue(t, []string{"c", "d", "c", "f"}),
				Unique: true,
			},
			original: `[a, c, c]`,
			expected: []string{"a", "c", "d", "f"},
		},
		{
			name:     "Valid Nop Map Merge",
			spec:     renderingPatchSpec{},
			original: `foo: bar`,
			expected: map[string]string{"foo": "bar"},
		},
		{
			name: "Valid Map Merge Only",
			spec: renderingPatchSpec{
				Merge: createMergeValue(t, map[string]string{
					"foo": "bar",
				}),
			},
			original: ``,
			expected: map[string]string{"foo": "bar"},
		},
		{
			name: "Map Merge No List Append",
			spec: renderingPatchSpec{
				Merge: createMergeValue(t, map[string][]string{
					"a": {"a"},
				}),
			},
			original: `a: [b, c]`,
			expected: map[string][]string{
				"a": {"a"},
			},
		},
		{
			name: "Map Merge Append List",
			spec: renderingPatchSpec{
				Merge: createMergeValue(t, map[string][]string{
					"a": {"a"},
				}),
				MapListAppend: true,
			},
			original: `a: [b, c]`,
			expected: map[string][]string{
				"a": {"b", "c", "a"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.spec.ApplyTo([]byte(test.original))
			if test.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			assert.Equal(t,
				createExpectedYamlValue(t, test.expected),
				string(result),
			)
		})
	}
}
