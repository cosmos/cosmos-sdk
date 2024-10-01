package testdata

import "cosmossdk.io/schema"

var ExampleSchema schema.ModuleSchema

var AllKindsObject schema.StateObjectType

func init() {
	AllKindsObject = schema.StateObjectType{
		Name: "all_kinds",
		KeyFields: []schema.Field{
			{
				Name: "id",
				Kind: schema.Int64Kind,
			},
			{
				Name: "ts",
				Kind: schema.TimeKind,
			},
		},
	}

	for i := schema.InvalidKind + 1; i <= schema.MAX_VALID_KIND; i++ {
		field := schema.Field{
			Name: i.String(),
			Kind: i,
		}

		switch i {
		case schema.EnumKind:
			field.ReferencedType = MyEnum.Name
		default:
		}

		AllKindsObject.ValueFields = append(AllKindsObject.ValueFields, field)
	}

	ExampleSchema = schema.MustCompileModuleSchema(
		AllKindsObject,
		SingletonObject,
		VoteObject,
		MyEnum,
		VoteType,
	)
}

var SingletonObject = schema.StateObjectType{
	Name: "singleton",
	ValueFields: []schema.Field{
		{
			Name: "foo",
			Kind: schema.StringKind,
		},
		{
			Name:     "bar",
			Kind:     schema.Int32Kind,
			Nullable: true,
		},
		{
			Name:           "an_enum",
			Kind:           schema.EnumKind,
			ReferencedType: MyEnum.Name,
		},
	},
}

var VoteObject = schema.StateObjectType{
	Name: "vote",
	KeyFields: []schema.Field{
		{
			Name: "proposal",
			Kind: schema.Int64Kind,
		},
		{
			Name: "address",
			Kind: schema.AddressKind,
		},
	},
	ValueFields: []schema.Field{
		{
			Name:           "vote",
			Kind:           schema.EnumKind,
			ReferencedType: VoteType.Name,
		},
	},
	RetainDeletions: true,
}

var VoteType = schema.EnumType{
	Name: "vote_type",
	Values: []schema.EnumValueDefinition{
		{Name: "yes", Value: 1},
		{Name: "no", Value: 2},
		{Name: "abstain", Value: 3},
	},
}

var MyEnum = schema.EnumType{
	Name: "my_enum",
	Values: []schema.EnumValueDefinition{
		{Name: "a", Value: 1},
		{Name: "b", Value: 2},
		{Name: "c", Value: 3},
	},
}
