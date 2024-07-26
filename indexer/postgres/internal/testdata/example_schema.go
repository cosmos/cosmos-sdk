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
			field.EnumType = MyEnum
		default:
		}

		AllKindsObject.ValueFields = append(AllKindsObject.ValueFields, field)
	}

	ExampleSchema = mustModuleSchema([]schema.ObjectType{
		AllKindsObject,
		SingletonObject,
		VoteObject,
	})
}

func mustModuleSchema(objectTypes []schema.ObjectType) schema.ModuleSchema {
	s, err := schema.NewModuleSchema(objectTypes)
	if err != nil {
		panic(err)
	}
	return s
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
			Name:     "an_enum",
			Kind:     schema.EnumKind,
			EnumType: MyEnum,
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
			Kind: schema.AddressKind,
		},
	},
	ValueFields: []schema.Field{
		{
			Name: "vote",
			Kind: schema.EnumKind,
			EnumType: schema.EnumType{
				Name:   "vote_type",
				Values: []string{"yes", "no", "abstain"},
			},
		},
	},
	RetainDeletions: true,
}

var MyEnum = schema.EnumType{
	Name:   "my_enum",
	Values: []string{"a", "b", "c"},
}
