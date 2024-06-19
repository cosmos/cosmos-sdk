package indexertesting

import (
	"fmt"

	indexerbase "cosmossdk.io/indexer/base"
)

var ExampleAppSchema = map[string]indexerbase.ModuleSchema{
	"all_kinds": mkAllKindsModule(),
	"test_cases": {
		ObjectTypes: []indexerbase.ObjectType{
			{
				"Singleton",
				[]indexerbase.Field{},
				[]indexerbase.Field{
					{
						Name: "Value",
						Kind: indexerbase.StringKind,
					},
					{
						Name: "Value2",
						Kind: indexerbase.BytesKind,
					},
				},
				false,
			},
			{
				Name: "Simple",
				KeyFields: []indexerbase.Field{
					{
						Name: "Key",
						Kind: indexerbase.StringKind,
					},
				},
				ValueFields: []indexerbase.Field{
					{
						Name: "Value1",
						Kind: indexerbase.Int32Kind,
					},
					{
						Name: "Value2",
						Kind: indexerbase.BytesKind,
					},
				},
			},
			{
				Name: "Two Keys",
				KeyFields: []indexerbase.Field{
					{
						Name: "Key1",
						Kind: indexerbase.StringKind,
					},
					{
						Name: "Key2",
						Kind: indexerbase.Int32Kind,
					},
				},
			},
			{
				Name: "Three Keys",
				KeyFields: []indexerbase.Field{
					{
						Name: "Key1",
						Kind: indexerbase.StringKind,
					},
					{
						Name: "Key2",
						Kind: indexerbase.Int32Kind,
					},
					{
						Name: "Key3",
						Kind: indexerbase.Uint64Kind,
					},
				},
				ValueFields: []indexerbase.Field{
					{
						Name: "Value1",
						Kind: indexerbase.Int32Kind,
					},
				},
			},
			{
				Name: "Many Values",
				KeyFields: []indexerbase.Field{
					{
						Name: "Key",
						Kind: indexerbase.StringKind,
					},
				},
				ValueFields: []indexerbase.Field{
					{
						Name: "Value1",
						Kind: indexerbase.Int32Kind,
					},
					{
						Name: "Value2",
						Kind: indexerbase.BytesKind,
					},
					{
						Name: "Value3",
						Kind: indexerbase.Float64Kind,
					},
					{
						Name: "Value4",
						Kind: indexerbase.Uint64Kind,
					},
				},
			},
			{
				Name: "RetainDeletions",
				KeyFields: []indexerbase.Field{
					{
						Name: "Key",
						Kind: indexerbase.StringKind,
					},
				},
				ValueFields: []indexerbase.Field{
					{
						Name: "Value1",
						Kind: indexerbase.Int32Kind,
					},
					{
						Name: "Value2",
						Kind: indexerbase.BytesKind,
					},
				},
				RetainDeletions: true,
			},
		},
	},
}

func mkAllKindsModule() indexerbase.ModuleSchema {
	mod := indexerbase.ModuleSchema{}

	for i := 1; i < int(indexerbase.MAX_VALID_KIND); i++ {
		kind := indexerbase.Kind(i)
		typ := mkTestObjectType(kind)
		mod.ObjectTypes = append(mod.ObjectTypes, typ)
	}

	return mod
}

func mkTestObjectType(kind indexerbase.Kind) indexerbase.ObjectType {
	field := indexerbase.Field{
		Kind: kind,
	}

	if kind == indexerbase.EnumKind {
		field.EnumDefinition = testEnum
	}

	if kind == indexerbase.Bech32AddressKind {
		field.AddressPrefix = "cosmos"
	}

	key1Field := field
	key1Field.Name = "keyNotNull"
	key2Field := field
	key2Field.Name = "keyNullable"
	key2Field.Nullable = true
	val1Field := field
	val1Field.Name = "valNotNull"
	val2Field := field
	val2Field.Name = "valNullable"
	val2Field.Nullable = true

	return indexerbase.ObjectType{
		Name:        fmt.Sprintf("test_%v", kind),
		KeyFields:   []indexerbase.Field{key1Field, key2Field},
		ValueFields: []indexerbase.Field{val1Field, val2Field},
	}
}

var testEnum = indexerbase.EnumDefinition{
	Name:   "test_enum",
	Values: []string{"foo", "bar", "baz"},
}
