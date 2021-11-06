package rs

import (
	"testing"

	"arhat.dev/pkg/testhelper"
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

func createPatchValue(t *testing.T, i interface{}) *yaml.Node {
	data, err := yaml.Marshal(i)
	if !assert.NoError(t, err) {
		t.FailNow()
		return nil
	}

	ret := new(yaml.Node)
	if !assert.NoError(t, yaml.Unmarshal(data, ret)) {
		t.FailNow()
		return nil
	}

	return ret
}

func createMergeValue(t *testing.T, i interface{}) []MergeSource {
	return []MergeSource{{Value: createPatchValue(t, i)}}
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

		spec PatchSpec

		expectErr bool
		expected  interface{}
	}{
		{
			name: "Valid Nop List Merge",
			spec: PatchSpec{
				Value: createPatchValue(t, []interface{}{"a", "b", "c"}),
			},
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name: "Valid List Merge Only",
			spec: PatchSpec{
				Value: nil,
				Merge: createMergeValue(t, []string{"a", "b", "c"}),
			},
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name: "Invalid List Merge Type Not Match",
			spec: PatchSpec{
				Value: createPatchValue(t, []interface{}{"a", "b", "c"}),
				Merge: createMergeValue(t, "oops: not a list"),
			},
			expectErr: true,
		},
		{
			name: "List Merge",
			spec: PatchSpec{
				Value: createPatchValue(t, []interface{}{"a", "b", "c"}),
				Merge: createMergeValue(t, []string{"c", "d", "e", "f"}),
			},
			expected: []interface{}{
				"a", "b", "c",
				"c", // expected dup
				"d", "e", "f",
			},
		},
		{
			name: "List Merge Unique",
			spec: PatchSpec{
				Value:  createPatchValue(t, []interface{}{"a", "c", "c"}),
				Merge:  createMergeValue(t, []string{"c", "d", "c", "f"}),
				Unique: true,
			},
			expected: []interface{}{"a", "c", "d", "f"},
		},
		{
			name: "Valid Nop Map Merge",
			spec: PatchSpec{
				Value: createPatchValue(t, map[string]interface{}{"foo": "bar"}),
			},
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name: "Valid Map Merge Only",
			spec: PatchSpec{
				Value: nil,
				Merge: createMergeValue(t, map[string]string{
					"foo": "bar",
				}),
			},
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name: "Map Merge No List Append",
			spec: PatchSpec{
				Value: createPatchValue(t, map[string]interface{}{"a": []interface{}{"b", "c"}}),
				Merge: createMergeValue(t, map[string][]string{
					"a": {"a"},
				}),
			},
			expected: map[string]interface{}{"a": []interface{}{"a"}},
		},
		{
			name: "Map Merge Append List",
			spec: PatchSpec{
				Value: createPatchValue(t, map[string]interface{}{"a": []interface{}{"b", "c"}}),
				Merge: createMergeValue(t, map[string][]string{
					"a": {"a"},
				}),
				MapListAppend: true,
			},
			expected: map[string]interface{}{
				"a": []interface{}{"b", "c", "a"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.spec.Apply(&testRenderingHandler{})
			if test.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, result)
		})
	}
}

func TestPatchSpec(t *testing.T) {
	testAnyObjectUnmarshalAndResolveByYamlSpecs(t, "testdata/patch-spec")
}

func TestPatchSpec_unresolved(t *testing.T) {
	type TestCase struct {
		BaseField

		Slice []interface{} `yaml:"slice"`
	}

	type Expected struct {
		Slice []interface{} `yaml:"slice"`
	}

	assertVisibleValues := func(t *testing.T, expected *Expected, actual *TestCase) {
		assert.EqualValues(t, expected.Slice, actual.Slice)
	}

	testhelper.TestFixtures(t, "./testdata/patch-spec-unresolved",
		func() interface{} { return Init(&TestCase{}, nil) },
		func() interface{} { return &Expected{} },
		func(t *testing.T, spec, exp interface{}) {
			in := spec.(*TestCase)
			expected := exp.(*Expected)

			assert.NoError(t, in.ResolveFields(&testRenderingHandler{}, -1))
			assertVisibleValues(t, expected, in)
		},
	)
}
