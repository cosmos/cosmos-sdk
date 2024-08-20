package schema

type EnumTypeDiff struct {
	Name          string
	AddedValues   []string
	RemovedValues []string
}

func DiffEnumTypes(oldEnum, newEnum EnumType) EnumTypeDiff {
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
