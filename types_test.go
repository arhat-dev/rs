package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var (
	_ RenderingHandler = RenderingHandleFunc(nil)

	_ InterfaceTypeHandler = InterfaceTypeHandleFunc(nil)
)

func TestInterfaceTypeHandleFunc_Create(t *testing.T) {
	called := false
	f := InterfaceTypeHandleFunc(func(typ reflect.Type, yamlKey string) (any, error) {
		called = true
		assert.EqualValues(t, "test", yamlKey)
		return nil, nil
	})

	f.Create(nil, "test")
	assert.True(t, called)
}

func TestRenderingHandler_RenderYaml(t *testing.T) {
	called := false
	f := RenderingHandleFunc(func(renderer string, rawData any) (result []byte, err error) {
		called = true
		assert.EqualValues(t, "test", renderer)
		assert.EqualValues(t, "rawData", rawData)
		return nil, nil
	})

	f.RenderYaml("test", "rawData")
	assert.True(t, called)
}

func TestNormalizeRawData(t *testing.T) {
	for _, test := range []struct {
		name     string
		rawData  any
		expected any
	}{
		{
			name: "str scalar node",
			rawData: func() *yaml.Node {
				n := new(yaml.Node)
				assert.NoError(t, yaml.Unmarshal([]byte("foo"), n))
				return prepareYamlNode(n)
			}(),
			expected: "foo",
		},
		{
			name: "binary scalar node",
			rawData: func() *yaml.Node {
				n := new(yaml.Node)
				assert.NoError(t, yaml.Unmarshal([]byte("!!binary Zm9v"), n))
				return prepareYamlNode(n)
			}(),
			expected: "foo",
		},
		{
			name: "null node",
			rawData: func() any {
				n := new(yaml.Node)
				assert.NoError(t, yaml.Unmarshal([]byte("null"), n))
				n = prepareYamlNode(n)
				assert.NotNil(t, n)
				return n
			}(),
			expected: nil,
		},
		{
			name:     "non node",
			rawData:  "foo",
			expected: "foo",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ret, err := NormalizeRawData(test.rawData)
			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, ret)
		})
	}
}
