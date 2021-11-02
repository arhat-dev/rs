package rs

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBaseField_HasUnresolvedField(t *testing.T) {
	f := &BaseField{}
	assert.False(t, f.HasUnresolvedField())

	f.addUnresolvedField("test", "test|data", nil, "foo", reflect.Value{}, false, nil)
	assert.True(t, f.HasUnresolvedField())
}

// TODO: remove this test once upstream issue solved
//
// issue: https://github.com/go-yaml/yaml/issues/665
func TestResolve_yaml_unmarshal_panic(t *testing.T) {
	tests := []struct {
		dataBytes string
	}{
		{"#\n- C\nD\n"},
	}

	for _, test := range tests {
		var out interface{}
		func() {
			defer func() {
				rec := recover()
				assert.NotNil(t, rec)
			}()

			err := yaml.Unmarshal([]byte(test.dataBytes), &out)
			assert.Error(t, fmt.Errorf("unreachable code: %w", err))
		}()
	}
}

func TestResolve_yaml_unmarshal_invalid_but_no_error(t *testing.T) {
	tests := []struct {
		dataBytes string
	}{
		{`[[]]test`},
	}

	for _, test := range tests {
		out := new(yaml.Node)
		err := yaml.Unmarshal([]byte(test.dataBytes), out)
		assert.NoError(t, err, "error return works?")

		md, err := yaml.Marshal(out)
		assert.NoError(t, err)
		assert.NotEqual(t, test.dataBytes, string(md))

		t.Log(string(md))
	}
}
