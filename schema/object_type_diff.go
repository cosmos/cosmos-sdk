package schema

type ObjectTypeDiff struct {
	Name            string
	KeyFieldsDiff   FieldsDiff
	ValueFieldsDiff FieldsDiff
}

type FieldsDiff struct {
	Added    []Field
	Changed  []FieldDiff
	Removed  []Field
	OldOrder []string
	NewOrder []string
}

func DiffObjectTypes(oldObj, newObj ObjectType) ObjectTypeDiff {
	diff := ObjectTypeDiff{
		Name: oldObj.TypeName(),
	}
	return diff
}

func DiffFields(oldFields, newFields []Field) FieldsDiff {
	diff := FieldsDiff{}

	newFieldMap := make(map[string]Field)
	for _, f := range newFields {
		newFieldMap[f.Name] = f
	}

	oldFieldMap := make(map[string]Field)
	for _, oldField := range oldFields {
		oldFieldMap[oldField.Name] = oldField
		newField, ok := newFieldMap[oldField.Name]
		if !ok {
			diff.Removed = append(diff.Removed, oldField)
		} else {
			fieldDiff := DiffField(oldField, newField)
			if !fieldDiff.Empty() {
				diff.Changed = append(diff.Changed, fieldDiff)
			}
		}
	}

	for _, newField := range newFields {
		if _, ok := oldFieldMap[newField.Name]; !ok {
			diff.Added = append(diff.Added, newField)
		}
	}

	oldOrder := make([]string, 0, len(oldFields))
	for _, f := range oldFields {
		oldOrder = append(oldOrder, f.Name)
	}

	orderChanged := false
	newOrder := make([]string, 0, len(newFields))
	for i, f := range newFields {
		newOrder = append(newOrder, f.Name)
		if i < len(oldOrder) && f.Name != oldOrder[i] {
			orderChanged = true
		}
	}

	if orderChanged {
		diff.OldOrder = oldOrder
		diff.NewOrder = newOrder
	}

	return diff
}

func (o ObjectTypeDiff) Empty() bool {
	return o.KeyFieldsDiff.Empty() && o.ValueFieldsDiff.Empty()
}

func (d FieldsDiff) Empty() bool {
	if len(d.Added) != 0 || len(d.Changed) != 0 || len(d.Removed) != 0 {
		return false
	}

	return !d.OrderChanged()
}

func (d FieldsDiff) OrderChanged() bool {
	if len(d.OldOrder) == 0 && len(d.NewOrder) == 0 {
		return false
	}

	return true
}
