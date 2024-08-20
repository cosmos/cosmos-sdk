package schema

import (
	"encoding/json"
	"reflect"
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

func TestFieldJSON(t *testing.T) {
	tt := []struct {
		field     Field
		json      string
		expectErr bool
	}{
		{
			field: Field{
				Name: "field1",
				Kind: StringKind,
			},
			json: `{"name":"field1","kind":"string"}`,
		},
		{
			field: Field{
				Name:     "field1",
				Kind:     Int32Kind,
				Nullable: true,
			},
			json: `{"name":"field1","kind":"int32","nullable":true}`,
		},
		{
			field: Field{
				Name: "field1",
				Kind: EnumKind,
				EnumType: EnumType{
					Name:   "enum",
					Values: []string{"a", "b"},
				},
			},
			json: `{"name":"field1","kind":"enum","enum_type":{"name":"enum","values":["a","b"]}}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.json, func(t *testing.T) {
			b, err := json.Marshal(tc.field)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if string(b) != tc.json {
					t.Fatalf("expected %s, got %s", tc.json, string(b))
				}
				var field Field
				err = json.Unmarshal(b, &field)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(field, tc.field) {
					t.Fatalf("expected %v, got %v", tc.field, field)
				}
			}
		})
	}
}
