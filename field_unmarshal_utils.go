package rs

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// alterInterface is a direct `interface{}` replacement for data unmarshaling
// with no rendering suffix support
type alterInterface struct {
	mapData   map[string]*alterInterface
	sliceData []*alterInterface

	scalarData interface{}

	rawScalarData string
}

func (f *alterInterface) HasValue() bool {
	return f.mapData != nil || f.sliceData != nil || f.scalarData != nil
}

func (f *alterInterface) NormalizedValue() interface{} {
	switch {
	case f.mapData != nil:
		ret := make(map[string]interface{}, len(f.mapData))
		for k, v := range f.mapData {
			ret[k] = v.NormalizedValue()
		}

		return ret
	case f.sliceData != nil:
		ret := make([]interface{}, len(f.sliceData))
		for i := range f.sliceData {
			ret[i] = f.sliceData[i].NormalizedValue()
		}

		return ret
	default:
		return f.scalarData
	}
}

func (f *alterInterface) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		switch n.ShortTag() {
		case "!!str":
			f.scalarData = n.Value
		case "!!binary":
			f.scalarData = n.Value
		default:
			f.rawScalarData = n.Value
			return n.Decode(&f.scalarData)
		}

		return nil
	case yaml.MappingNode:
		f.mapData = make(map[string]*alterInterface)
		return n.Decode(&f.mapData)
	case yaml.SequenceNode:
		f.sliceData = make([]*alterInterface, 0)
		return n.Decode(&f.sliceData)
	default:
		return fmt.Errorf("unexpected node tag %q", n.ShortTag())
	}
}

func (f *alterInterface) MarshalYAML() (interface{}, error) {
	switch {
	case f.mapData != nil:
		return f.mapData, nil
	case f.sliceData != nil:
		return f.sliceData, nil
	case len(f.rawScalarData) != 0:
		return f.rawScalarData, nil
	default:
		return f.scalarData, nil
	}
}
