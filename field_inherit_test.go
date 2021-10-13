package rs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseField_Inherit(t *testing.T) {
	a := &BaseField{}
	b := &BaseField{}

	// TODO: add tests
	assert.NoError(t, a.Inherit(b))
}
