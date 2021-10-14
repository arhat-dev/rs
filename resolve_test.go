package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseField_HasUnresolvedField(t *testing.T) {
	f := &BaseField{}
	assert.False(t, f.HasUnresolvedField())

	f.addUnresolvedField("test", "test|data", "foo", reflect.Value{}, false, nil)
	assert.True(t, f.HasUnresolvedField())
}
