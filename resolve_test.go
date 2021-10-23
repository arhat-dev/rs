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

	f.addUnresolvedField("test", "test|data", "foo", reflect.Value{}, false, nil)
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
