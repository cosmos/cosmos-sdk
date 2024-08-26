package diff

import "cosmossdk.io/schema"

// EnumTypeDiff represents the difference between two enum types.
type EnumTypeDiff struct {
	// Name is the name of the enum type.
	Name string

	// AddedValues is a list of values that were added.
	AddedValues []string

	// RemovedValues is a list of values that were removed.
	RemovedValues []string
}

func compareEnumType(oldEnum, newEnum schema.EnumType) EnumTypeDiff {
	diff := EnumTypeDiff{
		Name: oldEnum.TypeName(),
	}

	newValues := make(map[string]struct{})
	for _, v := range newEnum.Values {
		newValues[v] = struct{}{}
	}

	oldValues := make(map[string]struct{})
	for _, v := range oldEnum.Values {
		oldValues[v] = struct{}{}
		if _, ok := newValues[v]; !ok {
			diff.RemovedValues = append(diff.RemovedValues, v)
		}
	}

	for _, v := range newEnum.Values {
		if _, ok := oldValues[v]; !ok {
			diff.AddedValues = append(diff.AddedValues, v)
		}
	}

	return diff
}

// Empty returns true if the enum type diff has no changes.
func (e EnumTypeDiff) Empty() bool {
	return len(e.AddedValues) == 0 && len(e.RemovedValues) == 0
}

// HasCompatibleChanges returns true if the diff contains only compatible changes.
// The only supported compatible change is adding values.
func (e EnumTypeDiff) HasCompatibleChanges() bool {
	return len(e.RemovedValues) == 0
}
