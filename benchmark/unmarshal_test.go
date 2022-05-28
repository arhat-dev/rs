package benchmark

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"arhat.dev/pkg/stringhelper"
	"arhat.dev/rs"
	goccy_yaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func runUnmarshalBenchmarkPlain(b *testing.B, src []byte, unmarshal func([]byte, any) error) {
	var (
		f   PlainFoo
		err error
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = yaml.Unmarshal(src, &f); err != nil {
			b.Log(err)
			b.Fail()
		}

		if f.Str != "a" || f.Float != 10.1 || f.Bar.Map["m"] != "m" {
			b.Fail()
		}
	}
}

func runUnmarshalBenchmarkField(b *testing.B, src []byte, unmarshal func([]byte, any) error) {
	var (
		f   FieldFoo
		err error
	)

	rc := rs.RenderingHandleFunc(func(renderer string, rawData any) (result []byte, err error) {
		nr, err := rs.NormalizeRawData(rawData)
		switch t := nr.(type) {
		case string:
			return stringhelper.ToBytes[byte, byte](t), nil
		case []byte:
			return t, nil
		default:
			return yaml.Marshal(t)
		}
	})

	_ = rs.Init(&f, &rs.Options{AllowUnknownFields: true})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = yaml.Unmarshal(src, &f); err != nil {
			b.Log(err)
			b.Fail()
		}

		if err = f.ResolveFields(rc, -1); err != nil {
			b.Log(err)
			b.Fail()
		}

		if f.Str != "a" || f.Float != 10.1 || f.Bar.Map["m"] != "m" {
			b.Fail()
		}
	}
}

func BenchmarkUnmarshal_typed(b *testing.B) {
	srcNoRS := []byte(`{ str: a, float: 10.1, map: { m: m } }`)
	srcAllRS := []byte(`{ str@echo: a, float@echo: 10.1, map@echo: { m: m } }`)

	b.Run("go-yaml", func(b *testing.B) {
		runUnmarshalBenchmarkPlain(b, srcNoRS, yaml.Unmarshal)
	})

	b.Run("goccy-yaml", func(b *testing.B) {
		runUnmarshalBenchmarkPlain(b, srcNoRS, goccy_yaml.Unmarshal)
	})

	b.Run("rs", func(b *testing.B) {
		b.Run("no-suffix", func(b *testing.B) {
			runUnmarshalBenchmarkField(b, srcNoRS, yaml.Unmarshal)
		})

		b.Run("all-suffix", func(b *testing.B) {
			runUnmarshalBenchmarkField(b, srcAllRS, yaml.Unmarshal)
		})
	})
}

func BenchmarkUnmarshal_untyped(b *testing.B) {
	filepath.WalkDir("testdata", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		specYaml, err := os.ReadFile(path)
		if !assert.NoError(b, err, "failed to read test spec %q", d.Name()) {
			return err
		}

		b.Run(d.Name(), func(b *testing.B) {
			b.Run("json", func(b *testing.B) {
				var out any

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := json.Unmarshal(specYaml, &out); err != nil {
						b.Fail()
					}
				}
			})

			b.Run("go-yaml", func(b *testing.B) {
				var out any

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := yaml.Unmarshal(specYaml, &out); err != nil {
						b.Fail()
					}
				}
			})

			// goccy-yaml hangs forever
			// b.Run("goccy-yaml", func(b *testing.B) {
			// 	var out any
			// 	b.ResetTimer()
			// 	for i := 0; i < b.N; i++ {
			// 		if err := goccy_yaml.Unmarshal(specYaml, &out); err != nil {
			// 			b.Fail()
			// 		}
			// 	}
			// })

			b.Run("rs", func(b *testing.B) {
				var out rs.AnyObject

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := yaml.Unmarshal(specYaml, &out); err != nil {
						b.Fail()
					}
				}
			})
		})
		return nil
	})
}
