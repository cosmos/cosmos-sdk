package indexerbase

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
			name: "invalid module schema",
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
			errContains: "object type name cannot be empty",
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
