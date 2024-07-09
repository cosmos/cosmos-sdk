package testdata

import "cosmossdk.io/schema"

var ExampleSchema schema.ModuleSchema

var AllKindsObject schema.ObjectType

func init() {
	AllKindsObject = schema.ObjectType{
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
			field.EnumDefinition = MyEnum
		case schema.Bech32AddressKind:
			field.AddressPrefix = "foo"
		default:
		}

		AllKindsObject.ValueFields = append(AllKindsObject.ValueFields, field)
	}

	ExampleSchema = schema.ModuleSchema{
		ObjectTypes: []schema.ObjectType{
			AllKindsObject,
			SingletonObject,
			VoteObject,
		},
	}
}

var SingletonObject = schema.ObjectType{
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
			EnumDefinition: MyEnum,
		},
	},
}

var VoteObject = schema.ObjectType{
	Name: "vote",
	KeyFields: []schema.Field{
		{
			Name: "proposal",
			Kind: schema.Int64Kind,
		},
		{
			Name: "address",
			Kind: schema.Bech32AddressKind,
		},
	},
	ValueFields: []schema.Field{
		{
			Name: "vote",
			Kind: schema.EnumKind,
			EnumDefinition: schema.EnumDefinition{
				Name:   "vote_type",
				Values: []string{"yes", "no", "abstain"},
			},
		},
	},
	RetainDeletions: true,
}

var MyEnum = schema.EnumDefinition{
	Name:   "my_enum",
	Values: []string{"a", "b", "c"},
}
