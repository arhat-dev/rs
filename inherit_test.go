package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseField_Inherit(t *testing.T) {
	// TODO: add tests
	t.Run("both-empty", func(t *testing.T) {
		type Foo struct {
			BaseField

			Data string `yaml:"data"`
		}

		a := Init(&Foo{}, nil).(*Foo)
		b := Init(&Foo{}, nil).(*Foo)
		assert.NoError(t, a.Inherit(&b.BaseField))
	})

	t.Run("empty-inherit-some", func(t *testing.T) {
		type Foo struct {
			BaseField

			Data string `yaml:"data"`
		}

		empty := Init(&Foo{}, nil).(*Foo)
		some := Init(&Foo{}, nil).(*Foo)
		some.addUnresolvedField("data", "test", "Data",
			some._parentValue.FieldByName("Data"), false,
			&alterInterface{scalarData: "test-data"},
		)

		assert.NoError(t, empty.Inherit(&some.BaseField))

		assert.NotEqualValues(t, some._parentValue, empty._parentValue)
		assert.EqualValues(t,
			empty._parentValue.FieldByName("Data"),
			empty.unresolvedFields[unresolvedFieldKey{yamlKey: "data"}].fieldValue,
		)

		empty.unresolvedFields[unresolvedFieldKey{yamlKey: "data"}].fieldValue = some._parentValue.FieldByName("Data")
		some._parentValue = empty._parentValue
		assert.EqualValues(t, empty, some)
	})

	t.Run("some-inherit-some", func(t *testing.T) {
		type Foo struct {
			BaseField

			Data string `yaml:"data"`
		}

		a := Init(&Foo{}, nil).(*Foo)
		a.addUnresolvedField("data", "test", "Data",
			a._parentValue.FieldByName("Data"), false,
			&alterInterface{scalarData: "test-data"})
		b := Init(&Foo{}, nil).(*Foo)
		b.addUnresolvedField("data", "test", "Data",
			b._parentValue.FieldByName("Data"), false,
			&alterInterface{scalarData: "test-data"},
		)

		assert.NoError(t, a.Inherit(&b.BaseField))

		assert.NotEqualValues(t, b._parentValue, a._parentValue)
		assert.EqualValues(t,
			a._parentValue.FieldByName("Data"),
			a.unresolvedFields[unresolvedFieldKey{yamlKey: "data"}].fieldValue,
		)

		// set subtle values in a as they are in b
		a.unresolvedFields[unresolvedFieldKey{yamlKey: "data"}].fieldValue = b._parentValue.FieldByName("Data")
		b._parentValue = a._parentValue
		b.unresolvedFields[unresolvedFieldKey{
			yamlKey: "data",
		}].rawDataList = []*alterInterface{
			{scalarData: "test-data"},
			{scalarData: "test-data"},
		}
		assert.EqualValues(t, a, b)
	})

	t.Run("bad-inherit", func(t *testing.T) {
	})

	t.Run("catch-other-empty-inherit-some-with-cache", func(t *testing.T) {
		type Foo struct {
			BaseField

			Data map[string]string `rs:"other"`
		}

		empty := Init(&Foo{}, nil).(*Foo)
		some := Init(&Foo{}, nil).(*Foo)
		some.addUnresolvedField("data", "test", "Data",
			some._parentValue.FieldByName("Data"), true,
			&alterInterface{
				mapData: map[string]*alterInterface{
					"data": {scalarData: "test-data"},
				},
			},
		)
		some.catchOtherCache = map[string]reflect.Value{
			"a": reflect.ValueOf("cache-value"),
		}

		assert.NoError(t, empty.Inherit(&some.BaseField))

		assert.NotEqualValues(t, some._parentValue, empty._parentValue)
		assert.EqualValues(t,
			empty._parentValue.FieldByName("Data"),
			empty.unresolvedFields[unresolvedFieldKey{yamlKey: "data"}].fieldValue,
		)

		empty.unresolvedFields[unresolvedFieldKey{yamlKey: "data"}].fieldValue = some._parentValue.FieldByName("Data")
		some._parentValue = empty._parentValue
		assert.EqualValues(t, empty, some)
	})

	t.Run("catch-other-some-inherit-some-no-cache", func(t *testing.T) {
		type Foo struct {
			BaseField

			Data map[string]string `rs:"other"`
		}

		a := Init(&Foo{}, nil).(*Foo)
		a.addUnresolvedField("a", "test", "Data",
			a._parentValue.FieldByName("Data"), true,
			&alterInterface{
				mapData: map[string]*alterInterface{
					"a": {scalarData: "test-data"},
				},
			},
		)

		b := Init(&Foo{}, nil).(*Foo)
		b.addUnresolvedField("b", "test", "Data",
			b._parentValue.FieldByName("Data"), true,
			&alterInterface{
				mapData: map[string]*alterInterface{
					"b": {scalarData: "test-data"},
				},
			},
		)

		assert.NoError(t, a.Inherit(&b.BaseField))

		assert.NotEqualValues(t, b._parentValue, a._parentValue)
		assert.EqualValues(t,
			a._parentValue.FieldByName("Data"),
			a.unresolvedFields[unresolvedFieldKey{yamlKey: "b"}].fieldValue,
		)

		// set subtle values in a as they are in b
		a.unresolvedFields[unresolvedFieldKey{yamlKey: "a"}].fieldValue = b._parentValue.FieldByName("Data")
		a.unresolvedFields[unresolvedFieldKey{yamlKey: "b"}].fieldValue = b._parentValue.FieldByName("Data")
		b.addUnresolvedField("a", "test", "Data",
			b._parentValue.FieldByName("Data"), true,
			&alterInterface{
				mapData: map[string]*alterInterface{
					"a": {scalarData: "test-data"},
				},
			},
		)
		b._parentValue = a._parentValue

		assert.EqualValues(t, a, b)
	})
}
