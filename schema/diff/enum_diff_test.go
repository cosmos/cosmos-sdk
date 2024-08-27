package diff

import (
	"reflect"
	"testing"

	"cosmossdk.io/schema"
)

func Test_compareEnumType(t *testing.T) {
	tt := []struct {
		name                 string
		oldEnum              schema.EnumType
		newEnum              schema.EnumType
		diff                 EnumTypeDiff
		hasCompatibleChanges bool
	}{
		{
			name: "no change",
			oldEnum: schema.EnumType{
				Values: []schema.EnumValueDefinition{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
			},
			newEnum: schema.EnumType{
				Values: []schema.EnumValueDefinition{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
			},
			diff:                 EnumTypeDiff{},
			hasCompatibleChanges: true,
		},
		{
			name: "value added",
			oldEnum: schema.EnumType{
				Values: []schema.EnumValueDefinition{{Name: "a", Value: 1}},
			},
			newEnum: schema.EnumType{
				Values: []schema.EnumValueDefinition{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
			},
			diff: EnumTypeDiff{
				AddedValues: []schema.EnumValueDefinition{{Name: "b", Value: 2}},
			},
			hasCompatibleChanges: true,
		},
		{
			name: "value removed",
			oldEnum: schema.EnumType{
				Values: []schema.EnumValueDefinition{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
			},
			newEnum: schema.EnumType{
				Values: []schema.EnumValueDefinition{{Name: "a", Value: 1}},
			},
			diff: EnumTypeDiff{
				RemovedValues: []schema.EnumValueDefinition{{Name: "b", Value: 2}},
			},
			hasCompatibleChanges: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := compareEnumType(tc.oldEnum, tc.newEnum)
			if !reflect.DeepEqual(got, tc.diff) {
				t.Errorf("compareEnumType() = %v, want %v", got, tc.diff)
			}
			hasCompatibleChanges := got.HasCompatibleChanges()
			if hasCompatibleChanges != tc.hasCompatibleChanges {
				t.Errorf("HasCompatibleChanges() = %v, want %v", hasCompatibleChanges, tc.hasCompatibleChanges)
			}
		})
	}
}
