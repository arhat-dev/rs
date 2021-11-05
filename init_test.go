package rs

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func TestOptions_AllowedRenderers(t *testing.T) {
	tests := []struct {
		name      string
		allowList map[string]struct{}
		input     string
		expectErr bool
	}{
		{
			name:      "Nil - Permit All",
			allowList: nil,
			input:     "foo: bar",
			expectErr: false,
		},
		{
			name:      "Nil - Permit All",
			allowList: nil,
			input:     "foo@test|any: bar",
			expectErr: false,
		},
		{
			name:      "Empty - Deny All Good",
			allowList: map[string]struct{}{},
			input:     "foo: bar",
			expectErr: false,
		},
		{
			name:      "Empty - Deny All Bad",
			allowList: map[string]struct{}{},
			input:     "{foo: {foo: { foo@test|any: bar }}}",
			expectErr: true,
		},
		{
			name:      "Empty - Deny All Bad",
			allowList: map[string]struct{}{},
			input:     "[{ foo@test|any: bar }]",
			expectErr: true,
		},
		{
			name: "Non Empty - Permit Listed Good",
			allowList: map[string]struct{}{
				"test": {},
				"any":  {},
			},
			input:     "foo@test|any: bar",
			expectErr: false,
		},
		{
			name: "Non Empty - Permit Listed Good",
			allowList: map[string]struct{}{
				"": {},
			},
			input:     "foo@: bar",
			expectErr: false,
		},
		{
			name: "Non Empty - Permit Listed Bad",
			allowList: map[string]struct{}{
				"test": {},
			},
			input:     "foo@test|any: bar",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Init(&AnyObject{}, &Options{
				AllowedRenderers: test.allowList,
			})

			err := yaml.Unmarshal([]byte(test.input), out)
			if test.expectErr {
				assert.Error(t, err, fmt.Sprint(err))
			} else {
				assert.NoError(t, err, fmt.Sprint(err))
			}
		})
	}
}
