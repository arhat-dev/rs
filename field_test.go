package rs

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var _ RenderingHandler = (*testRenderingHandler)(nil)

type testRenderingHandler struct{}

func (h *testRenderingHandler) RenderYaml(renderer string, data interface{}) (result []byte, err error) {
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
	filepath.WalkDir(specDirPath, func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		specYaml, err := os.ReadFile(path)
		if !assert.NoError(t, err, "failed to read test spec %q", d.Name()) {
			return err
		}

		t.Run(d.Name(), func(t *testing.T) {
			dec := yaml.NewDecoder(bytes.NewReader(specYaml))

			var (
				testSrc  AnyObject
				expected interface{}
			)

			assert.NoError(t, dec.Decode(&testSrc), "invalid test source")
			assert.NoError(t, dec.Decode(&expected), "invalid expected result")

			assert.NoError(t,
				testSrc.ResolveFields(&testRenderingHandler{}, -1),
				"failed to resolve test source",
			)

			ret, err := yaml.Marshal(&testSrc)
			assert.NoError(t, err, "failed to marshal resolved data")

			var actual interface{}
			assert.NoError(t, yaml.Unmarshal(ret, &actual), "failed to unmarshal resolved data")

			assert.EqualValues(t, expected, actual)
		})

		return nil
	})
}
