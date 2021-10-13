package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseField_HasUnresolvedField(t *testing.T) {
	f := &BaseField{}
	f.addUnresolvedField("test", "test|data", "foo", reflect.Value{}, false, nil)
	assert.True(t, f.HasUnresolvedField())
}

func TestApplyTypeHint(t *testing.T) {
	tests := []struct {
		name     string
		hint     string
		value    interface{}
		expected interface{}

		expectErr bool
	}{
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
			name:      "Objects from map",
			hint:      "[]obj",
			value:     map[string]string{"foo": "bar"},
			expectErr: true,
		},

		//
		// Hint `map`
		//
		{
			name:     "Map from string",
			hint:     "map",
			value:    "foo: bar",
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "Map from bytes",
			hint:     "map",
			value:    []byte("foo: bar"),
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:      "Map from slice",
			hint:      "map",
			value:     []string{"foo", "bar"},
			expectErr: true,
		},
		{
			name:     "Map from map",
			hint:     "map",
			value:    map[string]string{"foo": "bar"},
			expected: map[string]string{"foo": "bar"},
		},

		//
		// Hint `int`
		//
		{
			name:     "Int from string",
			hint:     "int",
			value:    "10",
			expected: int(10),
		},
		{
			name:     "Int from bytes",
			hint:     "int",
			value:    []byte(`"10"`),
			expected: int(10),
		},
		{
			name:     "Int from int",
			hint:     "int",
			value:    int64(10),
			expected: int64(10),
		},
		{
			name:      "Int from slice",
			hint:      "int",
			value:     []string{"foo", "bar"},
			expectErr: true,
		},
		{
			name:      "Int from map",
			hint:      "int",
			value:     map[string]string{"foo": "bar"},
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

			ret, err := applyTypeHint(hint, &alterInterface{
				scalarData: test.value,
			})
			if test.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, ret)
		})
	}
}
