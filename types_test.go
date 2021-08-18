package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceTypeHandleFunc_Create(t *testing.T) {
	called := false
	f := InterfaceTypeHandleFunc(func(typ reflect.Type, yamlKey string) (interface{}, error) {
		called = true
		return nil, nil
	})

	f.Create(nil, "")
	assert.True(t, called)
}
