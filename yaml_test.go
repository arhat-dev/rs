package rs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestIsNull(t *testing.T) {
	for _, test := range []struct {
		input  string
		isNull bool
	}{
		{"", true},
		{"# all comments", true},
		{"map: val", false},
		{"1.1", false},
		{`"foo"`, false},
		{"null", true},
		{"~", true},
		{"[]", false},
	} {
		n := new(yaml.Node)
		assert.NoError(t, yaml.Unmarshal([]byte(test.input), n))
		assert.Equal(t, test.isNull, isNull(n), "test data: %q", test.input)
	}
}
