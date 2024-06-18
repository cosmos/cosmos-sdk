package indexerbase

import "testing"

func TestField_Validate(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		error string
	}{
		{
			name: "valid field",
			field: Field{
				Name: "field1",
				Kind: StringKind,
			},
			error: "",
		},
		{
			name: "empty name",
			field: Field{
				Name: "",
				Kind: StringKind,
			},
			error: "field name cannot be empty",
		},
		{
			name: "invalid kind",
			field: Field{
				Name: "field1",
				Kind: Kind(-1),
			},
			error: "invalid field type for \"field1\": invalid type: -1",
		},
		{
			name: "missing address prefix",
			field: Field{
				Name: "field1",
				Kind: Bech32AddressKind,
			},
			error: "missing address prefix for field \"field1\"",
		},
		{
			name: "address prefix with non-Bech32AddressKind",
			field: Field{
				Name:          "field1",
				Kind:          StringKind,
				AddressPrefix: "prefix",
			},
			error: "address prefix is only valid for field \"field1\" with type Bech32AddressKind",
		},
		{
			name: "invalid enum definition",
			field: Field{
				Name: "field1",
				Kind: EnumKind,
			},
			error: "invalid enum definition for field \"field1\": enum definition name cannot be empty",
		},
		{
			name: "enum definition with non-EnumKind",
			field: Field{
				Name:           "field1",
				Kind:           StringKind,
				EnumDefinition: EnumDefinition{Name: "enum"},
			},
			error: "enum definition is only valid for field \"field1\" with type EnumKind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.Validate()
			if tt.error == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if err.Error() != tt.error {
					t.Errorf("expected error: %s, got: %v", tt.error, err)
				}
			}
		})
	}
}
