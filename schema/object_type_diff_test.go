package schema

import (
	"reflect"
	"testing"
)

func Test_objectTypeDiff(t *testing.T) {
	tt := []struct {
		name                 string
		oldType              ObjectType
		newType              ObjectType
		diff                 ObjectTypeDiff
		trueF                func(ObjectTypeDiff) bool
		hasCompatibleChanges bool
	}{
		{
			name: "no change",
			oldType: ObjectType{
				KeyFields: []Field{{Name: "id", Kind: Int32Kind}},
			},
			newType: ObjectType{
				KeyFields: []Field{{Name: "id", Kind: Int32Kind}},
			},
			diff:                 ObjectTypeDiff{},
			trueF:                ObjectTypeDiff.Empty,
			hasCompatibleChanges: true,
		},
		{
			name: "key fields changed",
			oldType: ObjectType{
				KeyFields: []Field{{Name: "id", Kind: Int32Kind}},
			},
			newType: ObjectType{
				KeyFields: []Field{{Name: "id", Kind: StringKind}},
			},
			diff: ObjectTypeDiff{
				KeyFieldsDiff: FieldsDiff{
					Changed: []FieldDiff{
						{
							Name:    "id",
							OldKind: Int32Kind,
							NewKind: StringKind,
						},
					},
				},
			},
			trueF:                func(d ObjectTypeDiff) bool { return !d.KeyFieldsDiff.Empty() },
			hasCompatibleChanges: false,
		},
		{
			name: "value fields changed",
			oldType: ObjectType{
				ValueFields: []Field{{Name: "name", Kind: StringKind}},
			},
			newType: ObjectType{
				ValueFields: []Field{{Name: "name", Kind: Int32Kind}},
			},
			diff: ObjectTypeDiff{
				ValueFieldsDiff: FieldsDiff{
					Changed: []FieldDiff{
						{
							Name:    "name",
							OldKind: StringKind,
							NewKind: Int32Kind,
						},
					},
				},
			},
			trueF:                func(d ObjectTypeDiff) bool { return !d.ValueFieldsDiff.Empty() },
			hasCompatibleChanges: false,
		},
		{
			name: "value fields added",
			oldType: ObjectType{
				ValueFields: []Field{{Name: "id", Kind: Int32Kind}},
			},
			newType: ObjectType{
				ValueFields: []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: StringKind}},
			},
			diff: ObjectTypeDiff{
				ValueFieldsDiff: FieldsDiff{
					Added: []Field{{Name: "name", Kind: StringKind}},
				},
			},
			trueF:                func(d ObjectTypeDiff) bool { return !d.ValueFieldsDiff.Empty() },
			hasCompatibleChanges: true,
		},
		{
			name: "fields reordered",
			oldType: ObjectType{
				KeyFields:   []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: StringKind}},
				ValueFields: []Field{{Name: "x", Kind: Int32Kind}, {Name: "y", Kind: StringKind}},
			},
			newType: ObjectType{
				KeyFields:   []Field{{Name: "name", Kind: StringKind}, {Name: "id", Kind: Int32Kind}},
				ValueFields: []Field{{Name: "y", Kind: StringKind}, {Name: "x", Kind: Int32Kind}},
			},
			diff: ObjectTypeDiff{
				KeyFieldsDiff: FieldsDiff{
					OldOrder: []string{"id", "name"},
					NewOrder: []string{"name", "id"},
				},
				ValueFieldsDiff: FieldsDiff{
					OldOrder: []string{"x", "y"},
					NewOrder: []string{"y", "x"},
				},
			},
			trueF:                func(d ObjectTypeDiff) bool { return !d.KeyFieldsDiff.Empty() && !d.ValueFieldsDiff.Empty() },
			hasCompatibleChanges: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := compareObjectType(tc.oldType, tc.newType)
			if !reflect.DeepEqual(got, tc.diff) {
				t.Errorf("compareObjectType() = %v, want %v", got, tc.diff)
			}
			hasCompatibleChanges := got.HasCompatibleChanges()
			if hasCompatibleChanges != tc.hasCompatibleChanges {
				t.Errorf("HasCompatibleChanges() = %v, want %v", hasCompatibleChanges, tc.hasCompatibleChanges)
			}
		})
	}
}

func Test_fieldsDiff(t *testing.T) {
	tt := []struct {
		name      string
		oldFields []Field
		newFields []Field
		diff      FieldsDiff
	}{
		{
			name:      "no change",
			oldFields: []Field{{Name: "id", Kind: Int32Kind}},
			newFields: []Field{{Name: "id", Kind: Int32Kind}},
		},
		{
			name:      "field added",
			oldFields: []Field{{Name: "id", Kind: Int32Kind}},
			newFields: []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: StringKind}},
			diff: FieldsDiff{
				Added: []Field{{Name: "name", Kind: StringKind}},
			},
		},
		{
			name:      "field removed",
			oldFields: []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: StringKind}},
			newFields: []Field{{Name: "id", Kind: Int32Kind}},
			diff: FieldsDiff{
				Removed: []Field{{Name: "name", Kind: StringKind}},
			},
		},
		{
			name:      "field changed",
			oldFields: []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: StringKind}},
			newFields: []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: Int32Kind}},
			diff: FieldsDiff{
				Changed: []FieldDiff{
					{
						Name:    "name",
						OldKind: StringKind,
						NewKind: Int32Kind,
					},
				},
			},
		},
		{
			name:      "field order changed",
			oldFields: []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: StringKind}},
			newFields: []Field{{Name: "name", Kind: StringKind}, {Name: "id", Kind: Int32Kind}},
			diff: FieldsDiff{
				OldOrder: []string{"id", "name"},
				NewOrder: []string{"name", "id"},
			},
		},
		{
			name:      "field order changed with added fields",
			oldFields: []Field{{Name: "id", Kind: Int32Kind}},
			newFields: []Field{{Name: "name", Kind: StringKind}, {Name: "id", Kind: Int32Kind}},
			diff: FieldsDiff{
				Added:    []Field{{Name: "name", Kind: StringKind}},
				OldOrder: []string{"id"},
				NewOrder: []string{"name", "id"},
			},
		},
		{
			name:      "field order changed with removed fields",
			oldFields: []Field{{Name: "name", Kind: StringKind}, {Name: "id", Kind: Int32Kind}},
			newFields: []Field{{Name: "id", Kind: Int32Kind}},
			diff: FieldsDiff{
				Removed:  []Field{{Name: "name", Kind: StringKind}},
				OldOrder: []string{"name", "id"},
				NewOrder: []string{"id"},
			},
		},
		{
			name:      "field order changed with changed fields",
			oldFields: []Field{{Name: "name", Kind: StringKind}, {Name: "id", Kind: Int32Kind}},
			newFields: []Field{{Name: "id", Kind: Int32Kind}, {Name: "name", Kind: Int32Kind}},
			diff: FieldsDiff{
				Changed: []FieldDiff{
					{
						Name:    "name",
						OldKind: StringKind,
						NewKind: Int32Kind,
					},
				},
				OldOrder: []string{"name", "id"},
				NewOrder: []string{"id", "name"},
			},
		},
		{
			name: "added, removed, changed and reordered fields",
			oldFields: []Field{
				{Name: "id", Kind: Int32Kind},
				{Name: "name", Kind: StringKind},
				{Name: "age", Kind: Int32Kind},
			},
			newFields: []Field{
				{Name: "name", Kind: Int32Kind},
				{Name: "age", Kind: Int32Kind},
				{Name: "email", Kind: StringKind},
			},
			diff: FieldsDiff{
				Added:   []Field{{Name: "email", Kind: StringKind}},
				Removed: []Field{{Name: "id", Kind: Int32Kind}},
				Changed: []FieldDiff{
					{
						Name:    "name",
						OldKind: StringKind,
						NewKind: Int32Kind,
					},
				},
				OldOrder: []string{"id", "name", "age"},
				NewOrder: []string{"name", "age", "email"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := compareFields(tc.oldFields, tc.newFields)
			if !reflect.DeepEqual(got, tc.diff) {
				t.Errorf("compareFields() = %v, want %v", got, tc.diff)
			}
		})
	}
}
