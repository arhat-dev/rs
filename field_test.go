package rs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func (f *fieldRef) stripBase() *fieldRef {
	f.base = nil
	return f
}

var _ RenderingHandler = (*testRenderingHandler)(nil)

type testRenderingHandler struct{}

func (testRenderingHandler) RenderYaml(renderer string, data any) (result []byte, err error) {
	var dataBytes []byte
	switch vt := data.(type) {
	case string:
		dataBytes = []byte(vt)
	case []byte:
		dataBytes = vt
	case *yaml.Node:
		switch vt.ShortTag() {
		case strTag:
			dataBytes = []byte(vt.Value)
		case binaryTag:
			dataBytes = []byte(vt.Value)
		default:
			dataBytes, err = yaml.Marshal(data)
		}
	case nil:
	default:
		dataBytes, err = yaml.Marshal(vt)
	}

	switch renderer {
	case "echo":
		return dataBytes, err
	case "add-suffix-test":
		return append(dataBytes, "-test"...), err
	case "err":
		return nil, fmt.Errorf("always error")
	case "empty":
		return nil, nil
	}

	panic(fmt.Errorf("unexpected renderer name: %q", renderer))
}

func TestParseRenderingSuffix(t *testing.T) {
	for _, test := range []struct {
		suffix   string
		expected []rendererSpec
	}{
		{"", nil},
		{"|||||", nil},
		{"foo", []rendererSpec{{name: "foo"}}},
		{"foo!", []rendererSpec{{name: "foo", patchSpec: true}}},
		{"foo?int", []rendererSpec{{name: "foo", typeHint: TypeHintInt{}}}},
		{"foo?str!", []rendererSpec{{name: "foo", patchSpec: true, typeHint: TypeHintStr{}}}},
		{"foo|bar", []rendererSpec{{name: "foo"}, {name: "bar"}}},
	} {
		t.Run(test.suffix, func(t *testing.T) {
			ret := parseRenderingSuffix(test.suffix)
			assert.EqualValues(t, test.expected, ret)
		})
	}
}
