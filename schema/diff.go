package schema

type ModuleSchemaDiff struct {
	AddedObjectTypes   []ObjectType
	ChangedObjectTypes []ObjectTypeDiff
	RemovedObjectTypes []ObjectType
	AddedEnumTypes     []EnumType
	ChangedEnumTypes   []EnumTypeDiff
	RemovedEnumTypes   []EnumType
}

func DiffModuleSchemas(oldSchema, newSchema *ModuleSchema) ModuleSchemaDiff {
	diff := ModuleSchemaDiff{}

	oldObjectTypes := map[string]ObjectType{}
	oldSchema.ObjectTypes(func(oldObj ObjectType) bool {
		newTyp, ok := newSchema.LookupType(oldObj.TypeName())
		if !ok {
			diff.RemovedObjectTypes = append(diff.RemovedObjectTypes, oldObj)
			return true
		}
		newObj, ok := newTyp.(ObjectType)
		if !ok {
			diff.RemovedObjectTypes = append(diff.RemovedObjectTypes, oldObj)
			return true
		}
		objDiff := DiffObjectTypes(oldObj, newObj)
		if !objDiff.Empty() {
			diff.ChangedObjectTypes = append(diff.ChangedObjectTypes, objDiff)
		}
		return true
	})

	newSchema.ObjectTypes(func(newObj ObjectType) bool {
		_, ok := oldSchema.LookupType(newObj.TypeName())
		if !ok {
			diff.AddedObjectTypes = append(diff.AddedObjectTypes, newObj)
		}
		return true
	})

	oldEnumTypes := map[string]EnumType{}
	oldSchema.EnumTypes(func(oldEnum EnumType) bool {
		newTyp, ok := newSchema.LookupType(oldEnum.TypeName())
		if !ok {
			diff.RemovedEnumTypes = append(diff.RemovedEnumTypes, oldEnum)
			return true
		}
		newEnum, ok := newTyp.(EnumType)
		if !ok {
			diff.RemovedEnumTypes = append(diff.RemovedEnumTypes, oldEnum)
			return true
		}
		enumDiff := DiffEnumTypes(oldEnum, newEnum)
		if !enumDiff.Empty() {
			diff.ChangedEnumTypes = append(diff.ChangedEnumTypes, enumDiff)
		}
		return true
	})

	return diff
}
