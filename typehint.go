package rs

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type TypeHint interface {
	apply(*alterInterface) (*alterInterface, error)

	String() string
}

type (
	TypeHintNone    struct{}
	TypeHintStr     struct{}
	TypeHintBytes   struct{}
	TypeHintObject  struct{}
	TypeHintObjects struct{}
	TypeHintInt     struct{}
	TypeHintUint    struct{}
	TypeHintFloat   struct{}
)

func (TypeHintNone) String() string    { return "" }
func (TypeHintStr) String() string     { return "str" }
func (TypeHintBytes) String() string   { return "[]byte" }
func (TypeHintObject) String() string  { return "obj" }
func (TypeHintObjects) String() string { return "[]obj" }
func (TypeHintInt) String() string     { return "int" }
func (TypeHintUint) String() string    { return "uint" }
func (TypeHintFloat) String() string   { return "float" }

func (TypeHintNone) apply(v *alterInterface) (*alterInterface, error) {
	// no hint, use default behavior:
	// 	return string value if unmarshal failed or result
	//  is string or []byte

	var rawBytes []byte
	switch vt := v.scalarData.(type) {
	case string:
		rawBytes = []byte(vt)
	case []byte:
		rawBytes = vt
	default:
		return v, nil
	}

	tmp := new(alterInterface)
	err := yaml.Unmarshal(rawBytes, tmp)
	if err != nil {
		// couldn't unmarshal, return original value in string format
		return &alterInterface{
			scalarData: string(rawBytes),
		}, nil
	}

	switch tmp.Value().(type) {
	case string, []byte, nil:
		// yaml.Unmarshal will do some transformation on plaintext value
		// when it's not valid yaml, so return the original value
		return &alterInterface{
			scalarData: string(rawBytes),
		}, nil
	default:
		return tmp, nil
	}
}

func incompatibleTypeError(h TypeHint, v interface{}) error {
	return fmt.Errorf("typehint.%s: incompatible type %T", h, v)
}

func ParseTypeHint(h string) (TypeHint, error) {
	switch h {
	case "":
		return TypeHintNone{}, nil
	case "str":
		return TypeHintStr{}, nil
	case "[]byte":
		return TypeHintBytes{}, nil
	case "[]obj":
		return TypeHintObjects{}, nil
	case "obj":
		return TypeHintObject{}, nil
	case "int":
		return TypeHintInt{}, nil
	case "uint":
		return TypeHintUint{}, nil
	case "float":
		return TypeHintFloat{}, nil
	default:
		return nil, fmt.Errorf("unknown type hint %q", h)
	}
}
