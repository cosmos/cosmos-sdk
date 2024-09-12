package diff

import "cosmossdk.io/schema"

// ModuleSchemaDiff represents the difference between two module schemas.
type ModuleSchemaDiff struct {
	// AddedObjectTypes is a list of object types that were added.
	AddedObjectTypes []schema.StateObjectType

	// ChangedObjectTypes is a list of object types that were changed.
	ChangedObjectTypes []ObjectTypeDiff

	// RemovedObjectTypes is a list of object types that were removed.
	RemovedObjectTypes []schema.StateObjectType

	// AddedEnumTypes is a list of enum types that were added.
	AddedEnumTypes []schema.EnumType

	// ChangedEnumTypes is a list of enum types that were changed.
	ChangedEnumTypes []EnumTypeDiff

	// RemovedEnumTypes is a list of enum types that were removed.
	RemovedEnumTypes []schema.EnumType
}

// CompareModuleSchemas compares an old and a new module schemas and returns the difference between them.
// If the schemas are equivalent, the Empty method of the returned ModuleSchemaDiff will return true.
//
// Indexer implementations can use these diffs to perform automatic schema migration.
// The specific supported changes that a specific indexer supports are defined by that indexer implementation.
// However, as a general rule, it is suggested that indexers support the following changes to module schemas:
// - Adding object types
// - Adding enum types
// - Adding nullable value fields to object types
// - Adding enum values to enum types
//
// These changes are officially considered "compatible" changes, and the HasCompatibleChanges method of the returned
// ModuleSchemaDiff will return true if only compatible changes are present.
// Module authors can use the above guidelines as a reference point for what changes are generally
// considered safe to make to a module schema without breaking existing indexers.
func CompareModuleSchemas(oldSchema, newSchema schema.ModuleSchema) ModuleSchemaDiff {
	diff := ModuleSchemaDiff{}

	oldSchema.ObjectTypes(func(oldObj schema.ObjectType) bool {
		newObj, found := newSchema.LookupObjectType(oldObj.Name)
		if !found {
			diff.RemovedObjectTypes = append(diff.RemovedObjectTypes, oldObj)
			return true
		}
		objDiff := compareObjectType(oldObj, newObj)
		if !objDiff.Empty() {
			diff.ChangedObjectTypes = append(diff.ChangedObjectTypes, objDiff)
		}
		return true
	})

	newSchema.ObjectTypes(func(newObj schema.ObjectType) bool {
		_, found := oldSchema.LookupObjectType(newObj.TypeName())
		if !found {
			diff.AddedObjectTypes = append(diff.AddedObjectTypes, newObj)
		}
		return true
	})

	oldSchema.EnumTypes(func(oldEnum schema.EnumType) bool {
		newEnum, found := newSchema.LookupEnumType(oldEnum.Name)
		if !found {
			diff.RemovedEnumTypes = append(diff.RemovedEnumTypes, oldEnum)
			return true
		}
		enumDiff := compareEnumType(oldEnum, newEnum)
		if !enumDiff.Empty() {
			diff.ChangedEnumTypes = append(diff.ChangedEnumTypes, enumDiff)
		}
		return true
	})

	newSchema.EnumTypes(func(newEnum schema.EnumType) bool {
		_, found := oldSchema.LookupEnumType(newEnum.TypeName())
		if !found {
			diff.AddedEnumTypes = append(diff.AddedEnumTypes, newEnum)
		}
		return true
	})

	return diff
}

func (m ModuleSchemaDiff) Empty() bool {
	return len(m.AddedObjectTypes) == 0 &&
		len(m.ChangedObjectTypes) == 0 &&
		len(m.RemovedObjectTypes) == 0 &&
		len(m.AddedEnumTypes) == 0 &&
		len(m.ChangedEnumTypes) == 0 &&
		len(m.RemovedEnumTypes) == 0
}

// HasCompatibleChanges returns true if the diff contains only compatible changes.
// Compatible changes are changes that are generally safe to make to a module schema without breaking existing indexers
// and indexers should aim to automatically migrate to such changes.
// See the CompareModuleSchemas function for a list of changes that are considered compatible.
func (m ModuleSchemaDiff) HasCompatibleChanges() bool {
	// object and enum types can be added but not removed
	// changed object and enum types must have compatible changes
	if len(m.RemovedObjectTypes) != 0 || len(m.RemovedEnumTypes) != 0 {
		return false
	}

	for _, objectType := range m.ChangedObjectTypes {
		if !objectType.HasCompatibleChanges() {
			return false
		}
	}

	for _, enumType := range m.ChangedEnumTypes {
		if !enumType.HasCompatibleChanges() {
			return false
		}
	}

	return true
}
