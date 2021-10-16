package benchmark

import (
	"testing"

	"arhat.dev/rs"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func BenchmarkMarshal_typed(b *testing.B) {
	expected := map[string]interface{}{
		"str":   "a",
		"float": 10.1,
		"map": map[string]interface{}{
			"m": "m",
		},
	}
	b.Run("go-yaml", func(b *testing.B) {
		f := &PlainFoo{
			Str:   "a",
			Float: 10.1,
			Bar: struct{ Map map[string]string }{
				Map: map[string]string{"m": "m"},
			},
		}

		b.ResetTimer()

		var (
			out []byte
			err error
		)
		for i := 0; i < b.N; i++ {
			out, err = yaml.Marshal(f)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
		}
		b.StopTimer()

		actual := make(map[string]interface{})
		assert.NoError(b, yaml.Unmarshal(out, &actual))
		assert.EqualValues(b, expected, actual)
	})

	b.Run("goccy-yaml", func(b *testing.B) {
		f := &PlainFoo{
			Str:   "a",
			Float: 10.1,
			Bar: struct{ Map map[string]string }{
				Map: map[string]string{"m": "m"},
			},
		}

		b.ResetTimer()

		var (
			out []byte
			err error
		)
		for i := 0; i < b.N; i++ {
			out, err = yaml.Marshal(f)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
		}
		b.StopTimer()

		actual := make(map[string]interface{})
		assert.NoError(b, yaml.Unmarshal(out, &actual))
		assert.EqualValues(b, expected, actual)
	})

	b.Run("rs", func(b *testing.B) {
		f := rs.Init(&FieldFoo{
			Str:   "a",
			Float: 10.1,
			Bar: struct{ Map map[string]string }{
				Map: map[string]string{"m": "m"},
			},
		}, &rs.Options{AllowUnknownFields: true}).(*FieldFoo)

		b.ResetTimer()

		var (
			out []byte
			err error
		)
		for i := 0; i < b.N; i++ {
			out, err = yaml.Marshal(f)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
		}
		b.StopTimer()

		actual := make(map[string]interface{})
		assert.NoError(b, yaml.Unmarshal(out, &actual))
		assert.EqualValues(b, expected, actual)
	})
}

func BenchmarkMarshal_untyped(b *testing.B) {
	expected := map[string]interface{}{
		"str":   "a",
		"float": 10.1,
		"map": map[string]interface{}{
			"m": "m",
		},
	}

	b.Run("go-yaml", func(b *testing.B) {
		b.ResetTimer()

		var (
			out []byte
			err error
		)
		for i := 0; i < b.N; i++ {
			out, err = yaml.Marshal(expected)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
		}
		b.StopTimer()

		actual := make(map[string]interface{})
		assert.NoError(b, yaml.Unmarshal(out, &actual))
		assert.EqualValues(b, expected, actual)
	})

	b.Run("goccy-yaml", func(b *testing.B) {
		b.ResetTimer()

		var (
			out []byte
			err error
		)
		for i := 0; i < b.N; i++ {
			out, err = yaml.Marshal(expected)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
		}
		b.StopTimer()

		actual := make(map[string]interface{})
		assert.NoError(b, yaml.Unmarshal(out, &actual))
		assert.EqualValues(b, expected, actual)
	})

	b.Run("rs", func(b *testing.B) {
		f := rs.Init(&rs.AnyObjectMap{}, &rs.Options{AllowUnknownFields: true}).(*rs.AnyObjectMap)
		out, err := yaml.Marshal(expected)
		assert.NoError(b, err)
		assert.NoError(b, yaml.Unmarshal(out, f))

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			out, err = yaml.Marshal(f)
			if err != nil {
				b.Log(err)
				b.Fail()
			}
		}
		b.StopTimer()

		actual := make(map[string]interface{})
		assert.NoError(b, yaml.Unmarshal(out, &actual))
		assert.EqualValues(b, expected, actual)
	})
}
