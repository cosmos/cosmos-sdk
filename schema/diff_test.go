package schema

import (
	"reflect"
	"testing"
)

func TestCompareModuleSchemas(t *testing.T) {
	tt := []struct {
		name                 string
		oldSchema            ModuleSchema
		newSchema            ModuleSchema
		diff                 ModuleSchemaDiff
		hasCompatibleChanges bool
		empty                bool
	}{
		{
			name: "no change",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}},
			}),
			diff:                 ModuleSchemaDiff{},
			hasCompatibleChanges: true,
			empty:                true,
		},
		{
			name:      "object type added",
			oldSchema: mustModuleSchema(t),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}},
			}),
			diff: ModuleSchemaDiff{
				AddedObjectTypes: []ObjectType{
					{
						Name:      "object1",
						KeyFields: []Field{{Name: "key1", Kind: StringKind}},
					},
				},
			},
			hasCompatibleChanges: true,
		},
		{
			name: "object type removed",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}},
			}),
			newSchema: mustModuleSchema(t),
			diff: ModuleSchemaDiff{
				RemovedObjectTypes: []ObjectType{
					{
						Name:      "object1",
						KeyFields: []Field{{Name: "key1", Kind: StringKind}},
					},
				},
			},
			hasCompatibleChanges: false,
		},
		{
			name: "object type changed, key field added",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}, {Name: "key2", Kind: StringKind}},
			}),
			diff: ModuleSchemaDiff{
				ChangedObjectTypes: []ObjectTypeDiff{
					{
						Name: "object1",
						KeyFieldsDiff: FieldsDiff{
							Added: []Field{
								{Name: "key2", Kind: StringKind},
							},
						},
					},
				},
			},
			hasCompatibleChanges: false,
		},
		{
			name: "object type changed, nullable value field added",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:        "object1",
				KeyFields:   []Field{{Name: "key1", Kind: StringKind}},
				ValueFields: []Field{{Name: "value1", Kind: StringKind, Nullable: true}},
			}),
			diff: ModuleSchemaDiff{
				ChangedObjectTypes: []ObjectTypeDiff{
					{
						Name: "object1",
						ValueFieldsDiff: FieldsDiff{
							Added: []Field{{Name: "value1", Kind: StringKind, Nullable: true}},
						},
					},
				},
			},
			hasCompatibleChanges: true,
		},
		{
			name: "object type changed, non-nullable value field added",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:        "object1",
				KeyFields:   []Field{{Name: "key1", Kind: StringKind}},
				ValueFields: []Field{{Name: "value1", Kind: StringKind}},
			}),
			diff: ModuleSchemaDiff{
				ChangedObjectTypes: []ObjectTypeDiff{
					{
						Name: "object1",
						ValueFieldsDiff: FieldsDiff{
							Added: []Field{{Name: "value1", Kind: StringKind}},
						},
					},
				},
			},
			hasCompatibleChanges: false,
		},
		{
			name: "object type changed, fields reordered",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: StringKind}, {Name: "key2", Kind: StringKind}},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key2", Kind: StringKind}, {Name: "key1", Kind: StringKind}},
			}),
			diff: ModuleSchemaDiff{
				ChangedObjectTypes: []ObjectTypeDiff{
					{
						Name: "object1",
						KeyFieldsDiff: FieldsDiff{
							OldOrder: []string{"key1", "key2"},
							NewOrder: []string{"key2", "key1"},
						},
					},
				},
			},
			hasCompatibleChanges: false,
		},
		{
			name: "enum type added, nullable value field added",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: Int32Kind}},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: Int32Kind}},
				ValueFields: []Field{
					{
						Name:     "value1",
						Kind:     EnumKind,
						EnumType: EnumType{Name: "enum1", Values: []string{"a", "b"}},
						Nullable: true,
					},
				},
			}),
			diff: ModuleSchemaDiff{
				ChangedObjectTypes: []ObjectTypeDiff{
					{
						Name: "object1",
						ValueFieldsDiff: FieldsDiff{
							Added: []Field{
								{
									Name:     "value1",
									Kind:     EnumKind,
									EnumType: EnumType{Name: "enum1", Values: []string{"a", "b"}},
									Nullable: true,
								},
							},
						},
					},
				},
				AddedEnumTypes: []EnumType{
					{Name: "enum1", Values: []string{"a", "b"}},
				},
			},
			hasCompatibleChanges: true,
		},
		{
			name: "enum type removed",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: Int32Kind}},
				ValueFields: []Field{
					{
						Name:     "value1",
						Kind:     EnumKind,
						EnumType: EnumType{Name: "enum1", Values: []string{"a", "b"}},
					},
				},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: Int32Kind}},
			}),
			diff: ModuleSchemaDiff{
				ChangedObjectTypes: []ObjectTypeDiff{
					{
						Name: "object1",
						ValueFieldsDiff: FieldsDiff{
							Removed: []Field{
								{
									Name:     "value1",
									Kind:     EnumKind,
									EnumType: EnumType{Name: "enum1", Values: []string{"a", "b"}},
								},
							},
						},
					},
				},
				RemovedEnumTypes: []EnumType{
					{Name: "enum1", Values: []string{"a", "b"}},
				},
			},
			hasCompatibleChanges: false,
		},
		{
			name: "enum type value added",
			oldSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: EnumKind, EnumType: EnumType{Name: "enum1", Values: []string{"a"}}}},
			}),
			newSchema: mustModuleSchema(t, ObjectType{
				Name:      "object1",
				KeyFields: []Field{{Name: "key1", Kind: EnumKind, EnumType: EnumType{Name: "enum1", Values: []string{"a", "b"}}}},
			}),
			diff: ModuleSchemaDiff{
				ChangedEnumTypes: []EnumTypeDiff{
					{
						Name:        "enum1",
						AddedValues: []string{"b"},
					},
				},
			},
			hasCompatibleChanges: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := CompareModuleSchemas(tc.oldSchema, tc.newSchema)
			if !reflect.DeepEqual(got, tc.diff) {
				t.Errorf("CompareModuleSchemas() = %v, want %v", got, tc.diff)
			}
			hasCompatibleChanges := got.HasCompatibleChanges()
			if hasCompatibleChanges != tc.hasCompatibleChanges {
				t.Errorf("HasCompatibleChanges() = %v, want %v", hasCompatibleChanges, tc.hasCompatibleChanges)
			}
			if tc.empty != got.Empty() {
				t.Errorf("Empty() = %v, want %v", got.Empty(), tc.empty)
			}
		})
	}
}

func mustModuleSchema(t *testing.T, objectTypes ...ObjectType) ModuleSchema {
	s, err := NewModuleSchema(objectTypes)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
