package rs

import (
	"fmt"
	"path"
	"reflect"
	"testing"

	"arhat.dev/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBaseField_HasUnresolvedField(t *testing.T) {
	f := &BaseField{}
	assert.False(t, f.HasUnresolvedField())

	_ = f.addUnresolvedField("test", "test|data", nil,
		&fieldRef{
			fieldName:   "foo",
			fieldValue:  reflect.Value{},
			isInlineMap: false,
		},
		nil,
	)
	assert.True(t, f.HasUnresolvedField())
}

func TestBaseField_ResolveField(t *testing.T) {

	type ThirdLevelInput struct {
		BaseField

		Data interface{} `yaml:"data"`
	}

	type SecondLevelInput struct {
		BaseField

		L3   ThirdLevelInput `yaml:"l3"`
		Data interface{}     `yaml:"data"`
	}

	type TopLevelInput struct {
		BaseField

		L2   SecondLevelInput `yaml:"l2"`
		Data interface{}      `yaml:"data"`

		InlineMap map[string]interface{} `yaml:",inline"`
	}

	type TestCase struct {
		BaseField

		Input TopLevelInput `yaml:"input"`

		Depth           int      `yaml:"resolve_depth"`
		FieldsToResolve []string `yaml:"resolve_fields"`
	}

	type CheckSpec struct {
		BaseField

		Unmarshaled TopLevelInput `yaml:"unmarshaled"`
		Resolved    TopLevelInput `yaml:"resolved"`
	}

	assertVisibleValues := func(t *testing.T, expected, actual *TopLevelInput) bool {
		ret := assert.EqualValues(t, expected.Data, actual.Data)
		ret = ret && assert.EqualValues(t, expected.InlineMap, actual.InlineMap)

		ret = ret && assert.Equal(t, expected.L2.Data, actual.L2.Data)
		ret = ret && assert.Equal(t, expected.L2.L3.Data, actual.L2.L3.Data)
		return ret
	}

	testhelper.TestFixtures(t, "./testdata/resolve",
		func() interface{} { return Init(&TestCase{}, nil) },
		func() interface{} { return Init(&CheckSpec{}, nil) },
		func(t *testing.T, spec, exp interface{}) {
			input := spec.(*TestCase)
			expected := exp.(*CheckSpec)

			if !assertVisibleValues(t, &expected.Unmarshaled, &input.Input) {
				assert.Fail(t, "unmarshaled not match")
			}

			err := input.Input.ResolveFields(nil, input.Depth, input.FieldsToResolve...)
			assert.NoError(t, err)

			if !assertVisibleValues(t, &expected.Resolved, &input.Input) {
				assert.Fail(t, "resolved not match")
			}
		},
	)
}

// TODO: remove this test once upstream issue solved
//
// issue: https://github.com/go-yaml/yaml/issues/665
func TestResolve_yaml_unmarshal_panic(t *testing.T) {
	tests := []struct {
		dataBytes string
	}{
		{"#\n- C\nD\n"},
	}

	for _, test := range tests {
		var out interface{}
		func() {
			defer func() {
				rec := recover()
				assert.NotNil(t, rec)
			}()

			err := yaml.Unmarshal([]byte(test.dataBytes), &out)
			assert.Error(t, fmt.Errorf("unreachable code: %w", err))
		}()

		assert.Equal(t, test.dataBytes, assumeValidYaml([]byte(test.dataBytes)).Value)
	}
}

func TestResolve_yaml_unmarshal_invalid_but_no_error(t *testing.T) {
	tests := []struct {
		dataBytes string
	}{
		{`[[]]test`},
	}

	for _, test := range tests {
		out := new(yaml.Node)
		err := yaml.Unmarshal([]byte(test.dataBytes), out)
		assert.NoError(t, err, "error return works?")

		md, err := yaml.Marshal(out)
		assert.NoError(t, err)
		assert.NotEqual(t, test.dataBytes, string(md))

		t.Log(string(md))
	}
}

type testVirtualKeyItem struct {
	BaseField

	B string `yaml:"b"`
	A string `yaml:"a"`

	NestedObjects []*testVirtualKeyItem `yaml:"nested_objects"`
}

func assertTestVirtualKeyItemVisibleFields(t *testing.T, expected, actual *testVirtualKeyItem) {
	assert.EqualValues(t, expected.B, actual.B)
	assert.EqualValues(t, expected.A, actual.A)

	for i := range actual.NestedObjects {
		assertTestVirtualKeyItemVisibleFields(t, expected.NestedObjects[i], actual.NestedObjects[i])
	}
}

func TestVirtualKeyFixtures(t *testing.T) {
	type InlineMapObjects struct {
		BaseField

		InlineMap map[string][]*testVirtualKeyItem `yaml:",inline"`
	}

	type FooIface interface{}

	type InlineMapIfaceObjects struct {
		BaseField

		InlineMap map[string][]FooIface `yaml:",inline"`
	}

	type InlineMapObject struct {
		BaseField

		InlineMap map[string]*testVirtualKeyItem `yaml:",inline"`
	}

	type InlineMapIfaceObject struct {
		BaseField

		InlineMap map[string]FooIface `yaml:",inline"`
	}

	type TestSpec struct {
		BaseField

		Strings []string `yaml:"strings"`

		Objects []*testVirtualKeyItem `yaml:"objects"`

		InlineMap_Objects InlineMapObjects `yaml:"inline_map_objects"`

		InlineMap_IfaceObjects InlineMapIfaceObjects `yaml:"inline_map_iface_objects"`

		InlineMap_Object InlineMapObject `yaml:"inline_map_object"`

		InlineMap_IfaceObject InlineMapIfaceObject `yaml:"inline_map_iface_object"`
	}

	// 	type CheckSpec struct {
	// 		Strings []string `yaml:"strings"`
	//
	// 		InlineMapObjects struct {
	// 			InlineMap map[string][]map[string]interface{} `yaml:",inline"`
	// 		} `yaml:"inline_map_objects"`
	//
	// 		IfaceObjects struct {
	// 			InlineMap map[string][]map[string]interface{} `yaml:",inline"`
	// 		} `yaml:"inline_map_iface_objects"`
	// 	}

	opts := &Options{
		InterfaceTypeHandler: InterfaceTypeHandleFunc(
			func(typ reflect.Type, yamlKey string) (interface{}, error) {
				return &testVirtualKeyItem{}, nil
			},
		),
	}

	testhelper.TestFixtures(t, "./testdata/virtual-key",
		func() interface{} { return Init(&TestSpec{}, opts) },
		func() interface{} { return Init(&TestSpec{}, opts) },
		func(t *testing.T, in, exp interface{}) {
			actual := in.(*TestSpec)
			expected := exp.(*TestSpec)

			for i := 0; i < 5; i++ {
				t.Run(fmt.Sprint(i), func(t *testing.T) {
					name := path.Base(t.Name())
					_ = name
					err := actual.ResolveFields(&testRenderingHandler{}, -1)
					assert.NoError(t, err)
					assert.EqualValues(t, expected.Strings, actual.Strings)

					for i, e := range expected.Objects {
						a := expected.Objects[i]
						assertTestVirtualKeyItemVisibleFields(t, e, a)
					}

					for k, list := range expected.InlineMap_Objects.InlineMap {
						for i := range list {
							e := expected.InlineMap_Objects.InlineMap[k][i]
							a := actual.InlineMap_Objects.InlineMap[k][i]
							assertTestVirtualKeyItemVisibleFields(t, e, a)
						}
					}

					for k, list := range expected.InlineMap_IfaceObjects.InlineMap {
						for i := range list {
							e := expected.InlineMap_IfaceObjects.InlineMap[k][i].(*testVirtualKeyItem)
							a := actual.InlineMap_IfaceObjects.InlineMap[k][i].(*testVirtualKeyItem)
							assertTestVirtualKeyItemVisibleFields(t, e, a)
						}
					}

					for k := range expected.InlineMap_Object.InlineMap {
						e := expected.InlineMap_Object.InlineMap[k]
						a := actual.InlineMap_Object.InlineMap[k]
						assertTestVirtualKeyItemVisibleFields(t, e, a)
					}

					for k := range expected.InlineMap_IfaceObject.InlineMap {
						e := expected.InlineMap_IfaceObject.InlineMap[k].(*testVirtualKeyItem)
						a := actual.InlineMap_IfaceObject.InlineMap[k].(*testVirtualKeyItem)
						assertTestVirtualKeyItemVisibleFields(t, e, a)
					}
				})
			}
		},
	)
}
