package schema

import (
	"strings"
	"testing"
)

func TestModuleSchema_Validate(t *testing.T) {
	tests := []struct {
		name         string
		moduleSchema ModuleSchema
		errContains  string
	}{
		{
			name: "valid module schema",
			moduleSchema: ModuleSchema{
				ObjectTypes: []ObjectType{
					{
						Name: "object1",
						KeyFields: []Field{
							{
								Name: "field1",
								Kind: StringKind,
							},
						},
					},
				},
			},
			errContains: "",
		},
		{
			name: "invalid object type",
			moduleSchema: ModuleSchema{
				ObjectTypes: []ObjectType{
					{
						Name: "",
						KeyFields: []Field{
							{
								Name: "field1",
								Kind: StringKind,
							},
						},
					},
				},
			},
			errContains: "invalid object type name",
		},
		{
			name: "same enum with missing values",
			moduleSchema: ModuleSchema{
				ObjectTypes: []ObjectType{
					{
						Name: "object1",
						KeyFields: []Field{
							{
								Name: "k",
								Kind: EnumKind,
								EnumDefinition: EnumDefinition{
									Name:   "enum1",
									Values: []string{"a", "b"},
								},
							},
						},
						ValueFields: []Field{
							{
								Name: "v",
								Kind: EnumKind,
								EnumDefinition: EnumDefinition{
									Name:   "enum1",
									Values: []string{"a", "b", "c"},
								},
							},
						},
					},
				},
			},
			errContains: "different number of values",
		},
		{
			name: "same enum with different values",
			moduleSchema: ModuleSchema{
				ObjectTypes: []ObjectType{
					{
						Name: "object1",
						KeyFields: []Field{
							{
								Name: "k",
								Kind: EnumKind,
								EnumDefinition: EnumDefinition{
									Name:   "enum1",
									Values: []string{"a", "b"},
								},
							},
						},
					},
					{
						Name: "object2",
						KeyFields: []Field{
							{
								Name: "k",
								Kind: EnumKind,
								EnumDefinition: EnumDefinition{
									Name:   "enum1",
									Values: []string{"a", "c"},
								},
							},
						},
					},
				},
			},
			errContains: "different values",
		},
		{
			name: "same enum",
			moduleSchema: ModuleSchema{
				ObjectTypes: []ObjectType{
					{
						Name: "object1",
						KeyFields: []Field{
							{
								Name: "k",
								Kind: EnumKind,
								EnumDefinition: EnumDefinition{
									Name:   "enum1",
									Values: []string{"a", "b"},
								},
							},
						},
					},
					{
						Name: "object2",
						KeyFields: []Field{
							{
								Name: "k",
								Kind: EnumKind,
								EnumDefinition: EnumDefinition{
									Name:   "enum1",
									Values: []string{"a", "b"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.moduleSchema.Validate()
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestModuleSchema_ValidateObjectUpdate(t *testing.T) {
	tests := []struct {
		name         string
		moduleSchema ModuleSchema
		objectUpdate ObjectUpdate
		errContains  string
	}{
		{
			name: "valid object update",
			moduleSchema: ModuleSchema{
				ObjectTypes: []ObjectType{
					{
						Name: "object1",
						KeyFields: []Field{
							{
								Name: "field1",
								Kind: StringKind,
							},
						},
					},
				},
			},
			objectUpdate: ObjectUpdate{
				TypeName: "object1",
				Key:      "abc",
			},
			errContains: "",
		},
		{
			name: "object type not found",
			moduleSchema: ModuleSchema{
				ObjectTypes: []ObjectType{
					{
						Name: "object1",
						KeyFields: []Field{
							{
								Name: "field1",
								Kind: StringKind,
							},
						},
					},
				},
			},
			objectUpdate: ObjectUpdate{
				TypeName: "object2",
				Key:      "abc",
			},
			errContains: "object type \"object2\" not found in module schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.moduleSchema.ValidateObjectUpdate(tt.objectUpdate)
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}
