package rs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalYamlMap_merge_map(t *testing.T) {
	// editorconfig-checker-disable
	const input = `
other: &other
  a: b
  c:

foo:
bar: foo
<<: *other
`
	// editorconfig-checker-enable
	var out any
	assert.NoError(t, yaml.Unmarshal([]byte(input), &out))
	// assert.IsType(t, map[string]any{}, out)

	n := new(yaml.Node)
	assert.NoError(t, yaml.Unmarshal([]byte(input), n))
	n = prepareYamlNode(n)
	assert.EqualValues(t, 8, len(n.Content))

	pairs, err := unmarshalYamlMap(n.Content)
	assert.NoError(t, err)
	assert.EqualValues(t, len(pairs), 5)

	assert.Equal(t, "other", pairs[0][0].Value)
	assert.Equal(t, "", pairs[0][1].Value)
	assert.Equal(t, yaml.MappingNode, pairs[0][1].Kind)

	assert.Equal(t, "foo", pairs[1][0].Value)
	assert.Equal(t, "", pairs[1][1].Value)
	assert.Equal(t, yaml.ScalarNode, pairs[1][1].Kind)

	assert.Equal(t, "bar", pairs[2][0].Value)
	assert.Equal(t, "foo", pairs[2][1].Value)

	assert.Equal(t, "a", pairs[3][0].Value)
	assert.Equal(t, "b", pairs[3][1].Value)

	assert.Equal(t, "c", pairs[4][0].Value)
	assert.Equal(t, "", pairs[4][1].Value)
	assert.Equal(t, yaml.ScalarNode, pairs[4][1].Kind)
}
