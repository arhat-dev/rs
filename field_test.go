package rs

import (
	"fmt"
	"testing"

	"arhat.dev/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func (f *fieldRef) stripBase() *fieldRef {
	f.base = nil
	return f
}

var _ RenderingHandler = (*testRenderingHandler)(nil)

type testRenderingHandler struct{}

func (h *testRenderingHandler) RenderYaml(renderer string, data any) (result []byte, err error) {
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

	panic(fmt.Errorf("unexpected renderer name: %q", h))
}

func testAnyObjectUnmarshalAndResolveByYamlSpecs(t *testing.T, specDirPath string) {
	testhelper.TestFixtures(t, specDirPath,
		func() any { return &AnyObject{} },
		func() any { var out any; return &out },
		func(t *testing.T, spec, exp any) {
			testSrc := spec.(*AnyObject)
			exp = *exp.(*any)

			assert.NoError(t,
				testSrc.ResolveFields(&testRenderingHandler{}, -1),
				"failed to resolve test source",
			)

			ret, err := yaml.Marshal(&testSrc)
			assert.NoError(t, err, "failed to marshal resolved data")

			var actual any
			assert.NoError(t, yaml.Unmarshal(ret, &actual), "failed to unmarshal resolved data")

			assert.EqualValues(t, exp, actual)
		},
	)
}
