package rs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"arhat.dev/rs"
)

func TestLimitation_BuiltInMapKeyNotResolved(t *testing.T) {
	type Bar struct {
		rs.BaseField

		Data string `yaml:"data"`
	}

	type Foo struct {
		rs.BaseField

		BarMap map[string]*Bar `yaml:"bar_map"`
	}

	yamlData := `{ bar_map: { bar_key@my-renderer: { data: my-renderer-value } } }`
	f := rs.Init(&Foo{}, nil).(*Foo)
	assert.NoError(t, yaml.Unmarshal([]byte(yamlData), f))
	assert.NoError(t, f.ResolveFields(
		rs.RenderingHandleFunc(
			func(renderer string, rawData interface{}) (result interface{}, err error) {
				return rawData, nil
			},
		), -1),
	)

	assert.Nil(t, f.BarMap["bar_key"], "no such expected key if you want rendering suffix")
	assert.NotNil(t, f.BarMap["bar_key@my-renderer"], "only gives you raw yaml key")

	t.Run("workaround", func(t *testing.T) {
		type BarMap struct {
			rs.BaseField

			Data map[string]*Bar `rs:"other"`
		}

		type Foo struct {
			rs.BaseField

			BarMap BarMap `yaml:"bar_map"`
		}

		f := rs.Init(&Foo{}, nil).(*Foo)
		assert.NoError(t, yaml.Unmarshal([]byte(yamlData), f))
		assert.NoError(t, f.ResolveFields(
			rs.RenderingHandleFunc(
				func(renderer string, rawData interface{}) (result interface{}, err error) {
					return rawData, nil
				},
			), -1),
		)

		assert.NotNil(t, f.BarMap.Data["bar_key"], "now it works")
		assert.Nil(t, f.BarMap.Data["bar_key@my-renderer"])
	})
}
