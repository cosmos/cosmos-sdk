package schematesting

import (
	"fmt"

	"cosmossdk.io/schema"
)

// ExampleAppSchema is an example app schema that intends to cover all schema cases that indexers should handle
// that can be used in reproducible unit testing and property based testing.
var ExampleAppSchema = map[string]schema.ModuleSchema{
	"all_kinds": mkAllKindsModule(),
	"test_cases": schema.MustCompileModuleSchema(
		schema.StateObjectType{
			Name:      "Singleton",
			KeyFields: []schema.Field{},
			ValueFields: []schema.Field{
				{
					Name: "Value",
					Kind: schema.StringKind,
				},
				{
					Name: "Value2",
					Kind: schema.BytesKind,
				},
			},
		},
		schema.StateObjectType{
			Name: "Simple",
			KeyFields: []schema.Field{
				{
					Name: "Key",
					Kind: schema.StringKind,
				},
			},
			ValueFields: []schema.Field{
				{
					Name: "Value1",
					Kind: schema.Int32Kind,
				},
				{
					Name: "Value2",
					Kind: schema.BytesKind,
				},
			},
		},
		schema.StateObjectType{
			Name: "TwoKeys",
			KeyFields: []schema.Field{
				{
					Name: "Key1",
					Kind: schema.StringKind,
				},
				{
					Name: "Key2",
					Kind: schema.Int32Kind,
				},
			},
		},
		schema.StateObjectType{
			Name: "ThreeKeys",
			KeyFields: []schema.Field{
				{
					Name: "Key1",
					Kind: schema.StringKind,
				},
				{
					Name: "Key2",
					Kind: schema.Int32Kind,
				},
				{
					Name: "Key3",
					Kind: schema.Uint64Kind,
				},
			},
			ValueFields: []schema.Field{
				{
					Name: "Value1",
					Kind: schema.Int32Kind,
				},
			},
		},
		schema.StateObjectType{
			Name: "ManyValues",
			KeyFields: []schema.Field{
				{
					Name: "Key",
					Kind: schema.StringKind,
				},
			},
			ValueFields: []schema.Field{
				{
					Name: "Value1",
					Kind: schema.Int32Kind,
				},
				{
					Name: "Value2",
					Kind: schema.BytesKind,
				},
				{
					Name: "Value3",
					Kind: schema.Float64Kind,
				},
				{
					Name: "Value4",
					Kind: schema.Uint64Kind,
				},
			},
		},
		schema.StateObjectType{
			Name: "RetainDeletions",
			KeyFields: []schema.Field{
				{
					Name: "Key",
					Kind: schema.StringKind,
				},
			},
			ValueFields: []schema.Field{
				{
					Name: "Value1",
					Kind: schema.Int32Kind,
				},
				{
					Name: "Value2",
					Kind: schema.BytesKind,
				},
			},
			RetainDeletions: true,
		},
	),
}

func mkAllKindsModule() schema.ModuleSchema {
	types := []schema.Type{testEnum}
	for i := 1; i < int(schema.MAX_VALID_KIND); i++ {
		kind := schema.Kind(i)
		typ := mkTestObjectType(kind)
		types = append(types, typ)
	}

	return schema.MustCompileModuleSchema(types...)
}

func mkTestObjectType(kind schema.Kind) schema.StateObjectType {
	field := schema.Field{
		Kind: kind,
	}

	if kind == schema.EnumKind {
		field.ReferencedType = testEnum.Name
	}

	keyField := field
	keyField.Name = "key"
	if !kind.ValidKeyKind() {
		keyField.Kind = schema.Int32Kind
	}
	val1Field := field
	val1Field.Name = "valNotNull"
	val2Field := field
	val2Field.Name = "valNullable"
	val2Field.Nullable = true

	return schema.StateObjectType{
		Name:        fmt.Sprintf("test_%v", kind),
		KeyFields:   []schema.Field{keyField},
		ValueFields: []schema.Field{val1Field, val2Field},
	}
}

var testEnum = schema.EnumType{
	Name:   "test_enum_type",
	Values: []schema.EnumValueDefinition{{Name: "foo", Value: 1}, {Name: "bar", Value: 2}, {Name: "baz", Value: 3}},
}
