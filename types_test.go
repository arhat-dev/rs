package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	_ RenderingHandler = RenderingHandleFunc(nil)

	_ InterfaceTypeHandler = InterfaceTypeHandleFunc(nil)
)

func TestInterfaceTypeHandleFunc_Create(t *testing.T) {
	called := false
	f := InterfaceTypeHandleFunc(func(typ reflect.Type, yamlKey string) (interface{}, error) {
		called = true
		assert.EqualValues(t, "test", yamlKey)
		return nil, nil
	})

	f.Create(nil, "test")
	assert.True(t, called)
}

func TestRenderingHandler_RenderYaml(t *testing.T) {
	called := false
	f := RenderingHandleFunc(func(renderer string, rawData interface{}) (result interface{}, err error) {
		called = true
		assert.EqualValues(t, "test", renderer)
		assert.EqualValues(t, "rawData", rawData)
		return nil, nil
	})

	f.RenderYaml("test", "rawData")
	assert.True(t, called)
}
