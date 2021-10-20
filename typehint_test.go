package rs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type AlwaysErrorMarshaler struct{}

func (m *AlwaysErrorMarshaler) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("error for test purpose")
}

func TestTypeHint(t *testing.T) {
	tests := []struct {
		name     string
		hint     string
		value    interface{}
		expected interface{}

		expectErr bool
	}{
		// no hint
		{
			name:     "Default Str Map",
			value:    "{foo: bar}",
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "Default Str Map From Bytes",
			value:    []byte("{ foo: bar }"),
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "Default Str List",
			value:    "- foo\n- bar\n",
			expected: []interface{}{"foo", "bar"},
		},
		{
			name:     "Default Str",
			value:    "foo",
			expected: "foo",
		},
		{
			name:     "Default Str Can be malformed but preserved as is",
			value:    "# foo",
			expected: "# foo",
		},
		{
			name:     "Default Str from Invalid Yaml",
			value:    "[foo, bar, cool",
			expected: "[foo, bar, cool",
		},

		//
		// Hint `str`
		//
		{
			name:     "String from string",
			hint:     "str",
			value:    "foo",
			expected: "foo",
		},
		{
			// leave string as is
			name:     "String from invalid yaml",
			hint:     "str",
			value:    "{ foo ",
			expected: "{ foo ",
		},
		{
			name:     "String from bytes",
			hint:     "str",
			value:    []byte("foo"),
			expected: "foo",
		},
		{
			name:     "String from map",
			hint:     "str",
			value:    map[string]string{"foo": "bar"},
			expected: "foo: bar\n",
		},
		{
			name:     "String from slice",
			hint:     "str",
			value:    []string{"foo", "bar"},
			expected: "- foo\n- bar\n",
		},
		{
			name:      "String from invalid data",
			hint:      "str",
			value:     &AlwaysErrorMarshaler{},
			expectErr: true,
		},

		//
		// Hint `[]byte`
		//
		{
			name:     "Bytes from string",
			hint:     "[]byte",
			value:    "foo",
			expected: []byte("foo"),
		},
		{
			name:     "Bytes from bytes",
			hint:     "[]byte",
			value:    []byte("foo"),
			expected: []byte("foo"),
		},
		{
			name:     "Bytes from map",
			hint:     "[]byte",
			value:    map[string]string{"foo": "bar"},
			expected: []byte("foo: bar\n"),
		},
		{
			name:     "Bytes from slice",
			hint:     "[]byte",
			value:    []string{"foo", "bar"},
			expected: []byte("- foo\n- bar\n"),
		},
		{
			name:      "Bytes from invalid data",
			hint:      "[]byte",
			value:     &AlwaysErrorMarshaler{},
			expectErr: true,
		},

		//
		// Hint `obj`
		//
		{
			name:     "Object from string",
			hint:     "obj",
			value:    "foo: bar",
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "Object from bytes",
			hint:     "obj",
			value:    []byte("foo: bar"),
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:      "Object from malformed string",
			hint:      "obj",
			value:     "{ foo: bar",
			expectErr: true,
		},
		{
			name:      "Object from slice",
			hint:      "obj",
			value:     []string{"foo", "bar"},
			expectErr: true,
		},
		{
			name:     "Object from map",
			hint:     "obj",
			value:    map[string]string{"foo": "bar"},
			expected: map[string]string{"foo": "bar"},
		},
		{
			name: "Object from Struct Ptr",
			hint: "obj",

			value: func() ***struct{ A string } {
				v := struct{ A string }{A: "b"}
				vp := &v
				vpp := &vp
				return &vpp
			}(),

			expected: func() ***struct{ A string } {
				v := struct{ A string }{A: "b"}
				vp := &v
				vpp := &vp
				return &vpp
			}(),
		},

		{
			name: "Object from Non Ptr",
			hint: "obj",
			value: func() *int {
				i := 9
				return &i
			}(),
			expectErr: true,
		},

		//
		// Hint `[]obj`
		//
		{
			name:     "Objects from string",
			hint:     "[]obj",
			value:    "- foo\n- bar\n",
			expected: []interface{}{"foo", "bar"},
		},
		{
			name:      "Objects from malformed string",
			hint:      "[]obj",
			value:     "[ foo, bar",
			expectErr: true,
		},
		{
			name:     "Objects from bytes",
			hint:     "[]obj",
			value:    []byte("- foo\n- bar\n"),
			expected: []interface{}{"foo", "bar"},
		},
		{
			name:     "Objects from slice",
			hint:     "[]obj",
			value:    []string{"foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "Objects from slice Ptr",
			hint:     "[]obj",
			value:    &[]string{"foo", "bar"},
			expected: &[]string{"foo", "bar"},
		},
		{
			name:      "Objects from map",
			hint:      "[]obj",
			value:     map[string]string{"foo": "bar"},
			expectErr: true,
		},
		{
			name:      "Objects from invalid data",
			hint:      "[]obj",
			value:     &AlwaysErrorMarshaler{},
			expectErr: true,
		},

		//
		// Hint `int`
		//
		{
			name:     "Int from string",
			hint:     "int",
			value:    "-10",
			expected: int(-10),
		},
		{
			name:      "Int from malformed string",
			hint:      "int",
			value:     "-10.1",
			expectErr: true,
		},
		{
			name:     "Int from bytes",
			hint:     "int",
			value:    []byte(`"-10"`),
			expected: int(-10),
		},
		{
			name:     "Int from int",
			hint:     "int",
			value:    int64(-10),
			expected: int64(-10),
		},
		{
			name:     "Int from uint",
			hint:     "int",
			value:    uint64(10),
			expected: int(10),
		},
		{
			name:     "Int from float",
			hint:     "int",
			value:    float64(10.1),
			expected: int(10),
		},
		{
			name:      "Int from slice",
			hint:      "int",
			value:     []string{"foo", "bar"},
			expectErr: true,
		},

		//
		// Hint `uint`
		//
		{
			name:     "Uint from string",
			hint:     "uint",
			value:    "10",
			expected: uint(10),
		},
		{
			name:      "Uint from malformed string",
			hint:      "uint",
			value:     "-10",
			expectErr: true,
		},
		{
			name:     "Uint from bytes",
			hint:     "uint",
			value:    []byte(`"10"`),
			expected: uint(10),
		},
		{
			name:     "Uint from uint",
			hint:     "uint",
			value:    uint64(10),
			expected: uint64(10),
		},
		{
			name:     "Uint from int",
			hint:     "uint",
			value:    int(10),
			expected: uint(10),
		},
		{
			name:     "Uint from float",
			hint:     "uint",
			value:    float32(10.1),
			expected: uint(10),
		},
		{
			name:      "Uint from slice",
			hint:      "uint",
			value:     []string{"foo", "bar"},
			expectErr: true,
		},

		//
		// Hint `float`
		//
		{
			name:     "Float from string",
			hint:     "float",
			value:    "10.10",
			expected: float64(10.1),
		},
		{
			name:      "Float from malformed string",
			hint:      "float",
			value:     "1k",
			expectErr: true,
		},
		{
			name:     "Float from bytes",
			hint:     "float",
			value:    []byte(`"10.10"`),
			expected: float64(10.1),
		},
		{
			name:     "Float from int",
			hint:     "float",
			value:    int64(10),
			expected: float64(10),
		},
		{
			name:     "Float from uint",
			hint:     "float",
			value:    uint64(10),
			expected: float64(10),
		},
		{
			name:     "Float from float",
			hint:     "float",
			value:    float64(10.1),
			expected: float64(10.1),
		},
		{
			name:      "Float from slice",
			hint:      "float",
			value:     []string{"foo", "bar"},
			expectErr: true,
		},
		{
			name:      "Float from map",
			hint:      "float",
			value:     map[string]string{"foo": "bar"},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hint, err := ParseTypeHint(test.hint)
			assert.NoError(t, err)

			assert.EqualValues(t, test.hint, hint.String())

			ret, err := hint.apply(&alterInterface{
				scalarData: test.value,
			})
			if test.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, ret.NormalizedValue())
		})
	}
}

func TestParseTypeHint(t *testing.T) {
	t.Run("supported", func(t *testing.T) {
		for _, hint := range []string{
			"obj", "str", "[]byte", "[]obj", "int", "uint", "float",
		} {
			_, err := ParseTypeHint(hint)
			assert.NoError(t, err)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for _, hint := range []string{
			"a", "b",
		} {
			_, err := ParseTypeHint(hint)
			assert.Error(t, err)
		}
	})
}

func TestTypeHint_apply_crash(t *testing.T) {
	tests := []struct {
		name     string
		hint     TypeHint
		value    interface{}
		expected interface{}

		expectErr bool
	}{
		// no hint crash
		{
			name:     "Crash ",
			value:    "{foo: bar}",
			expected: map[string]interface{}{"foo": "bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}
