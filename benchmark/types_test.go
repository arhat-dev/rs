package benchmark

import "arhat.dev/rs"

type Bar struct {
	Map map[string]string `yaml:"map" json:"map"`
}

type PlainFoo struct {
	Str   string  `yaml:"str" json:"str"`
	Float float64 `yaml:"float" json:"float"`

	Bar `yaml:",inline" json:",inline"`
}

type FieldFoo struct {
	rs.BaseField

	Str   string  `yaml:"str"`
	Float float64 `yaml:"float"`

	Bar `yaml:",inline"`
}
