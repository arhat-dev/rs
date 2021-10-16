package benchmark

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"arhat.dev/rs"
	goccy_yaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type PlainFoo struct {
	Str   string
	Float float64

	Bar struct {
		Map map[string]string
	} `yaml:",inline"`
}

type FieldFoo struct {
	rs.BaseField

	Str   string
	Float float64

	Bar struct {
		Map map[string]string
	} `yaml:",inline"`
}

func BenchmarkUnmarshal_typed(b *testing.B) {
	srcNoRS := []byte(`{ str: a, float: 10.1, map: { m: m } }`)
	srcAllRS := []byte(`{ str@echo: a, float@echo: 10.1, map@echo: { m: m } }`)

	b.Run("go-yaml", func(b *testing.B) {
		f := &PlainFoo{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := yaml.Unmarshal(srcNoRS, f); err != nil {
				b.Log(err)
				b.Fail()
			}

			if f.Str != "a" || f.Float != 10.1 || f.Bar.Map["m"] != "m" {
				b.Fail()
			}
		}
	})

	b.Run("goccy-yaml", func(b *testing.B) {
		f := &PlainFoo{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := goccy_yaml.Unmarshal(srcNoRS, f); err != nil {
				b.Log(err)
				b.Fail()
			}

			if f.Str != "a" || f.Float != 10.1 || f.Bar.Map["m"] != "m" {
				b.Fail()
			}
		}
	})

	b.Run("rs", func(b *testing.B) {

		b.Run("no-suffix", func(b *testing.B) {
			f := rs.Init(&FieldFoo{}, &rs.Options{AllowUnknownFields: true}).(*FieldFoo)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := yaml.Unmarshal(srcNoRS, f); err != nil {
					b.Log(err)
					b.Fail()
				}

				if f.Str != "a" || f.Float != 10.1 || f.Bar.Map["m"] != "m" {
					b.Fail()
				}
			}
		})

		b.Run("all-suffix", func(b *testing.B) {
			f := rs.Init(&FieldFoo{}, &rs.Options{AllowUnknownFields: true}).(*FieldFoo)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := yaml.Unmarshal(srcAllRS, f); err != nil {
					b.Log(err)
					b.Fail()
				}

				if f.Str != "" || f.Float != 0 || f.Bar.Map["m"] != "" {
					b.Fail()
				}
			}
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
				var out interface{}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := json.Unmarshal(specYaml, &out); err != nil {
						b.Fail()
					}
				}
			})

			b.Run("go-yaml", func(b *testing.B) {
				var out interface{}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := yaml.Unmarshal(specYaml, &out); err != nil {
						b.Fail()
					}
				}
			})

			b.Run("goccy-yaml", func(b *testing.B) {
				// goccy-yaml hangs forever
				b.SkipNow()
				var out interface{}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := goccy_yaml.Unmarshal(specYaml, &out); err != nil {
						b.Fail()
					}
				}
			})

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
