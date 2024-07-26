package schema

import (
	"strings"
	"testing"
)

func TestField_Validate(t *testing.T) {
	tests := []struct {
		name        string
		field       Field
		errContains string
	}{
		{
			name: "valid field",
			field: Field{
				Name: "field1",
				Kind: StringKind,
			},
			errContains: "",
		},
		{
			name: "empty name",
			field: Field{
				Name: "",
				Kind: StringKind,
			},
			errContains: "invalid field name",
		},
		{
			name: "invalid kind",
			field: Field{
				Name: "field1",
				Kind: InvalidKind,
			},
			errContains: "invalid field kind",
		},
		{
			name: "invalid enum definition",
			field: Field{
				Name: "field1",
				Kind: EnumKind,
			},
			errContains: "invalid enum definition",
		},
		{
			name: "enum definition with non-EnumKind",
			field: Field{
				Name:     "field1",
				Kind:     StringKind,
				EnumType: EnumType{Name: "enum"},
			},
			errContains: "enum definition is only valid for field \"field1\" with type EnumKind",
		},
		{
			name: "valid enum",
			field: Field{
				Name:     "field1",
				Kind:     EnumKind,
				EnumType: EnumType{Name: "enum", Values: []string{"a", "b"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.Validate()
			if tt.errContains == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error contains: %s, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestField_ValidateValue(t *testing.T) {
	tests := []struct {
		name        string
		field       Field
		value       interface{}
		errContains string
	}{
		{
			name: "valid field",
			field: Field{
				Name: "field1",
				Kind: StringKind,
			},
			value:       "value",
			errContains: "",
		},
		{
			name: "null non-nullable field",
			field: Field{
				Name:     "field1",
				Kind:     StringKind,
				Nullable: false,
			},
			value:       nil,
			errContains: "cannot be null",
		},
		{
			name: "null nullable field",
			field: Field{
				Name:     "field1",
				Kind:     StringKind,
				Nullable: true,
			},
			value:       nil,
			errContains: "",
		},
		{
			name: "invalid value",
			field: Field{
				Name: "field1",
				Kind: StringKind,
			},
			value:       1,
			errContains: "invalid value for field \"field1\"",
		},
		{
			name: "valid enum",
			field: Field{
				Name:     "field1",
				Kind:     EnumKind,
				EnumType: EnumType{Name: "enum", Values: []string{"a", "b"}},
			},
			value:       "a",
			errContains: "",
		},
		{
			name: "invalid enum",
			field: Field{
				Name:     "field1",
				Kind:     EnumKind,
				EnumType: EnumType{Name: "enum", Values: []string{"a", "b"}},
			},
			value:       "c",
			errContains: "not a valid enum value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.ValidateValue(tt.value)
			if tt.errContains == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error contains: %s, got: %v", tt.errContains, err)
				}
			}
		})
	}
}
