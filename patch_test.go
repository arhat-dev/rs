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

		original   map[string]any
		additional map[string]any
		unique     bool

		expectErr bool
		expected  map[string]any
	}{
		{
			name:       "Simple Nop",
			original:   map[string]any{"foo": "bar"},
			additional: nil,
			expected:   map[string]any{"foo": "bar"},
		},
		{
			name:       "Simple Merge",
			original:   map[string]any{"foo": "bar"},
			additional: map[string]any{"foo": "bar"},
			expected:   map[string]any{"foo": "bar"},
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
	mapVal := map[string]any{
		"foo": "bar",
		"bar": map[string]any{
			"foo": "bar",
		},
	}
	tests := []struct {
		name string

		input    []any
		expected []any
	}{
		{
			name:     "Simple String",
			input:    []any{"a", "c", "c", "a"},
			expected: []any{"a", "c"},
		},
		{
			name:     "Simple Number",
			input:    []any{1, 1, 1, 1, 1},
			expected: []any{1},
		},
		{
			name:     "Map Value",
			input:    []any{mapVal, mapVal, 1},
			expected: []any{mapVal, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.EqualValues(t, test.expected, UniqueList(test.input))
		})
	}
}

func createPatchValue(t *testing.T, i any) *yaml.Node {
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

func createMergeValue(t *testing.T, i any) []MergeSource {
	return []MergeSource{{Value: createPatchValue(t, i)}}
}

func createExpectedYamlValue(t *testing.T, i any) string {
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
		expected  any
	}{
		{
			name: "Valid Nop List Merge",
			spec: PatchSpec{
				Value: createPatchValue(t, []any{"a", "b", "c"}),
			},
			expected: []any{"a", "b", "c"},
		},
		{
			name: "Valid List Merge Only",
			spec: PatchSpec{
				Value: nil,
				Merge: createMergeValue(t, []string{"a", "b", "c"}),
			},
			expected: []any{"a", "b", "c"},
		},
		{
			name: "Invalid List Merge Type Not Match",
			spec: PatchSpec{
				Value: createPatchValue(t, []any{"a", "b", "c"}),
				Merge: createMergeValue(t, "oops: not a list"),
			},
			expectErr: true,
		},
		{
			name: "List Merge",
			spec: PatchSpec{
				Value: createPatchValue(t, []any{"a", "b", "c"}),
				Merge: createMergeValue(t, []string{"c", "d", "e", "f"}),
			},
			expected: []any{
				"a", "b", "c",
				"c", // expected dup
				"d", "e", "f",
			},
		},
		{
			name: "List Merge Unique",
			spec: PatchSpec{
				Value:  createPatchValue(t, []any{"a", "c", "c"}),
				Merge:  createMergeValue(t, []string{"c", "d", "c", "f"}),
				Unique: true,
			},
			expected: []any{"a", "c", "d", "f"},
		},
		{
			name: "Valid Nop Map Merge",
			spec: PatchSpec{
				Value: createPatchValue(t, map[string]any{"foo": "bar"}),
			},
			expected: map[string]any{"foo": "bar"},
		},
		{
			name: "Valid Map Merge Only",
			spec: PatchSpec{
				Value: nil,
				Merge: createMergeValue(t, map[string]string{
					"foo": "bar",
				}),
			},
			expected: map[string]any{"foo": "bar"},
		},
		{
			name: "Map Merge No List Append",
			spec: PatchSpec{
				Value: createPatchValue(t, map[string]any{"a": []any{"b", "c"}}),
				Merge: createMergeValue(t, map[string][]string{
					"a": {"a"},
				}),
			},
			expected: map[string]any{"a": []any{"a"}},
		},
		{
			name: "Map Merge Append List",
			spec: PatchSpec{
				Value: createPatchValue(t, map[string]any{"a": []any{"b", "c"}}),
				Merge: createMergeValue(t, map[string][]string{
					"a": {"a"},
				}),
				MapListAppend: true,
			},
			expected: map[string]any{
				"a": []any{"b", "c", "a"},
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

		Slice []any `yaml:"slice"`
	}

	type Expected struct {
		Slice []any `yaml:"slice"`
	}

	assertVisibleValues := func(t *testing.T, expected *Expected, actual *TestCase) {
		assert.EqualValues(t, expected.Slice, actual.Slice)
	}

	testhelper.TestFixtures(t, "./testdata/patch-spec-unresolved",
		func() any { return Init(&TestCase{}, nil) },
		func() any { return &Expected{} },
		func(t *testing.T, spec, exp any) {
			in := spec.(*TestCase)
			expected := exp.(*Expected)

			assert.NoError(t, in.ResolveFields(testRenderingHandler{}, -1))
			assertVisibleValues(t, expected, in)
		},
	)
}
