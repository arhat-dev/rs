package rs_test

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"arhat.dev/rs"
)

func ExampleBaseField_ResolveFields() {
	type MyStruct struct {
		rs.BaseField `yaml:"-"`

		MyValue string `yaml:"my_value"`
	}

	// required:
	// initialize you data object
	// before marshaling/unmarshaling/resolving
	s := rs.Init(&MyStruct{}, nil)

	// unmarshal yaml data using rendering suffix
	err := yaml.Unmarshal([]byte(`{ my_value@my-renderer: 123 }`), s)
	if err != nil {
		panic(err)
	}

	err = s.ResolveFields(
		// implement your own renderer
		rs.RenderingHandleFunc(
			func(renderer string, rawData any) (result []byte, err error) {
				// usually you should have rawData normalized as golang types
				// so you don't have to tend to low level *yaml.Node objects
				rawData, err = rs.NormalizeRawData(rawData)
				if err != nil {
					return nil, err
				}

				switch vt := rawData.(type) {
				case string:
					return []byte(vt), nil
				case []byte:
					return vt, nil
				default:
					return []byte("hello"), nil
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
