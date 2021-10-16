package rs_test

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"arhat.dev/rs"
)

func ExampleField_ResolveFields() {
	type MyStruct struct {
		rs.BaseField `yaml:"-"`

		MyValue string `yaml:"my_value"`
	}

	// required:
	// initialize you data object
	// before marshaling/unmarshaling/resolving
	s := rs.Init(&MyStruct{}, nil).(*MyStruct)

	// unmarshal yaml data using rendering suffix
	err := yaml.Unmarshal([]byte(`{ my_value@my-renderer: 123 }`), s)
	if err != nil {
		panic(err)
	}

	err = s.ResolveFields(
		// implement your own renderer
		rs.RenderingHandleFunc(
			func(renderer string, rawData interface{}) (result interface{}, err error) {
				switch dt := rawData.(type) {
				case string:
					return dt, nil
				default:
					return "hello", nil
				}
			},
		),
		-1,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(s.MyValue)

	// output:
	//	hello
	//
}
