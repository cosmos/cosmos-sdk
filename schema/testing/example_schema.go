package schematesting

import (
	"fmt"

	"cosmossdk.io/schema"
)

// ExampleAppSchema is an example app schema that intends to cover all schema cases that indexers should handle
// that can be used in reproducible unit testing and property based testing.
var ExampleAppSchema = map[string]schema.ModuleSchema{
	"all_kinds": mkAllKindsModule(),
	"test_cases": MustNewModuleSchema([]schema.ObjectType{
		{
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
		{
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
		{
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
		{
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
		{
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
		{
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
	}),
}

func mkAllKindsModule() schema.ModuleSchema {
	var objTypes []schema.ObjectType
	for i := 1; i < int(schema.MAX_VALID_KIND); i++ {
		kind := schema.Kind(i)
		typ := mkTestObjectType(kind)
		objTypes = append(objTypes, typ)
	}

	return MustNewModuleSchema(objTypes)
}

func mkTestObjectType(kind schema.Kind) schema.ObjectType {
	field := schema.Field{
		Kind: kind,
	}

	if kind == schema.EnumKind {
		field.EnumType = testEnum
	}

	keyField := field
	keyField.Name = "key"
	val1Field := field
	val1Field.Name = "valNotNull"
	val2Field := field
	val2Field.Name = "valNullable"
	val2Field.Nullable = true

	return schema.ObjectType{
		Name:        fmt.Sprintf("test_%v", kind),
		KeyFields:   []schema.Field{keyField},
		ValueFields: []schema.Field{val1Field, val2Field},
	}
}

var testEnum = schema.EnumType{
	Name:   "test_enum_type",
	Values: []string{"foo", "bar", "baz"},
}
