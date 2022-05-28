package benchmark

import (
	"encoding/json"
	"testing"

	"arhat.dev/rs"
	goccy_yaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func runMarshalBenchmark[T any](
	b *testing.B,
	expected map[string]any,
	data T,
	marshal func(any) ([]byte, error),
	unmarshal func([]byte, any) error,
) {
	var (
		out []byte
		err error
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err = marshal(data)
		if err != nil {
			b.Log(err)
			b.Fail()
		}
	}
	b.StopTimer()

	actual := make(map[string]any)
	assert.NoError(b, unmarshal(out, &actual))
	assert.EqualValues(b, expected, actual)
}

func BenchmarkMarshal_typed(b *testing.B) {
	expected := map[string]any{
		"str":   "a",
		"float": 10.1,
		"map": map[string]any{
			"m": "m",
		},
	}

	b.Run("json", func(b *testing.B) {
		runMarshalBenchmark(b, expected, &PlainFoo{
			Str:   "a",
			Float: 10.1,
			Bar: Bar{
				Map: map[string]string{"m": "m"},
			},
		},
			json.Marshal,
			json.Unmarshal,
		)
	})

	b.Run("go-yaml", func(b *testing.B) {
		runMarshalBenchmark(b, expected, &PlainFoo{
			Str:   "a",
			Float: 10.1,
			Bar: Bar{
				Map: map[string]string{"m": "m"},
			},
		},
			yaml.Marshal,
			yaml.Unmarshal,
		)
	})

	b.Run("goccy-yaml", func(b *testing.B) {
		runMarshalBenchmark(b, expected, &PlainFoo{
			Str:   "a",
			Float: 10.1,
			Bar: Bar{
				Map: map[string]string{"m": "m"},
			},
		},
			goccy_yaml.Marshal,
			goccy_yaml.Unmarshal,
		)
	})

	b.Run("rs", func(b *testing.B) {
		runMarshalBenchmark(b, expected, rs.Init(&FieldFoo{
			Str:   "a",
			Float: 10.1,
			Bar: Bar{
				Map: map[string]string{"m": "m"},
			},
		}, &rs.Options{AllowUnknownFields: true}),
			yaml.Marshal,
			yaml.Unmarshal,
		)
	})
}

func BenchmarkMarshal_untyped(b *testing.B) {
	expected := map[string]any{
		"str":   "a",
		"float": 10.1,
		"map": map[string]any{
			"m": "m",
		},
	}

	b.Run("json", func(b *testing.B) {
		runMarshalBenchmark(b, expected, expected, json.Marshal, json.Unmarshal)
	})
	b.Run("go-yaml", func(b *testing.B) {
		runMarshalBenchmark(b, expected, expected, yaml.Marshal, yaml.Unmarshal)
	})
	b.Run("goccy-yaml", func(b *testing.B) {
		runMarshalBenchmark(b, expected, expected, goccy_yaml.Marshal, goccy_yaml.Unmarshal)
	})

	b.Run("rs", func(b *testing.B) {
		f := rs.Init(&rs.AnyObjectMap{}, &rs.Options{AllowUnknownFields: true})
		out, err := yaml.Marshal(expected)
		assert.NoError(b, err)
		assert.NoError(b, yaml.Unmarshal(out, f))

		runMarshalBenchmark(b, expected, f, yaml.Marshal, yaml.Unmarshal)
	})
}
