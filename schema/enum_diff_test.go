package schema

import (
	"reflect"
	"testing"
)

func Test_compareEnumType(t *testing.T) {
	tt := []struct {
		name                 string
		oldEnum              EnumType
		newEnum              EnumType
		diff                 EnumTypeDiff
		hasCompatibleChanges bool
	}{
		{
			name: "no change",
			oldEnum: EnumType{
				Values: []string{"a", "b"},
			},
			newEnum: EnumType{
				Values: []string{"a", "b"},
			},
			diff:                 EnumTypeDiff{},
			hasCompatibleChanges: true,
		},
		{
			name: "value added",
			oldEnum: EnumType{
				Values: []string{"a"},
			},
			newEnum: EnumType{
				Values: []string{"a", "b"},
			},
			diff: EnumTypeDiff{
				AddedValues: []string{"b"},
			},
			hasCompatibleChanges: true,
		},
		{
			name: "value removed",
			oldEnum: EnumType{
				Values: []string{"a", "b"},
			},
			newEnum: EnumType{
				Values: []string{"a"},
			},
			diff: EnumTypeDiff{
				RemovedValues: []string{"b"},
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
