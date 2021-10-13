package rs

import (
	"fmt"
)

var _ RenderingHandler = (*testRenderingHandler)(nil)

type testRenderingHandler struct{}

func (h *testRenderingHandler) RenderYaml(renderer string, data interface{}) (result interface{}, err error) {
	switch renderer {
	case "echo":
		return data, nil
	case "add-suffix-test":
		switch t := data.(type) {
		case string:
			return t + "-test", nil
		case []byte:
			return append(t, "-test"...), nil
		default:
			return nil, fmt.Errorf("unsupported non string nor bytes type: %T", data)
		}
	case "err":
		return nil, fmt.Errorf("always error")
	case "empty":
		return nil, nil
	}

	panic("unexpected renderer name")
}
