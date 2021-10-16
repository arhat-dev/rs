package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	type testFieldStruct struct {
		BaseField

		Foo string `yaml:"foo"`
	}

	var _ Field = (*testFieldStruct)(nil)

	fStruct := &testFieldStruct{}

	type testFieldPtr struct {
		*BaseField

		Foo string `yaml:"foo"`
	}

	fPtr1 := testFieldPtr{}
	fPtr2 := &testFieldPtr{}

	tests := []struct {
		name        string
		targetType  Field
		panicOnInit bool

		getBaseFieldParentValue func() reflect.Value
		getBaseFieldParentType  func() reflect.Type

		setDirectFoo          func(v string)
		getBaseFieldParentFoo func() string
	}{
		{
			name:       "Ptr BaseField",
			targetType: fStruct,
			getBaseFieldParentValue: func() reflect.Value {
				return fStruct.BaseField._parentValue
			},
			getBaseFieldParentType: func() reflect.Type {
				return fStruct._parentType
			},
			setDirectFoo: func(v string) {
				fStruct.Foo = v
			},
			getBaseFieldParentFoo: func() string {
				return fStruct.BaseField._parentValue.Interface().(testFieldStruct).Foo
			},
		},
		{
			name:        "Struct *BaseField",
			targetType:  fPtr1,
			panicOnInit: true,
		},
		{
			name:       "Ptr *BaseField",
			targetType: fPtr2,
			getBaseFieldParentValue: func() reflect.Value {
				return fPtr2.BaseField._parentValue
			},
			getBaseFieldParentType: func() reflect.Type {
				return fPtr2.BaseField._parentType
			},
			setDirectFoo: func(v string) {
				fPtr2.Foo = v
			},
			getBaseFieldParentFoo: func() string {
				return fPtr2.BaseField._parentValue.Interface().(testFieldPtr).Foo
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.panicOnInit {
				func() {
					defer func() {
						assert.NotNil(t, recover())
					}()

					_ = Init(test.targetType, nil)
				}()

				return
			}

			_ = Init(test.targetType, nil)

			assert.Equal(t, test.getBaseFieldParentValue().Type(), test.getBaseFieldParentType())
			assert.Equal(t, reflect.Struct, test.getBaseFieldParentValue().Kind())

			test.setDirectFoo("newValue")
			assert.Equal(t, "newValue", test.getBaseFieldParentFoo())
		})
	}
}
