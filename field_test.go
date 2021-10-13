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

func testUsingYamlSpecs(t *testing.T, specDirPath string) {
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
