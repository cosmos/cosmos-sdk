package schema

// ObjectTypeDiff represents the difference between two object types.
// The Empty method of KeyFieldsDiff and ValueFieldsDiff can be used to determine
// if there were any changes to the key fields or value fields.
type ObjectTypeDiff struct {
	// Name is the name of the object type.
	Name string

	// KeyFieldsDiff is the difference between the key fields of the object type.
	KeyFieldsDiff FieldsDiff

	// ValueFieldsDiff is the difference between the value fields of the object type.
	ValueFieldsDiff FieldsDiff
}

// FieldsDiff represents the difference between two lists of fields.
// Fields will be compared based on name first, and then if there is any
// difference in ordering that will be reported in OldOrder and NewOrder.
// If there is any order change, the OrderChanged method will return true.
// If fields were only added or removed but the order otherwise didn't change,
// then the OldOrder and NewOrder will still be empty.
type FieldsDiff struct {
	// Added is a list of fields that were added.
	Added []Field

	// Changed is a list of fields that were changed.
	Changed []FieldDiff

	// Removed is a list of fields that were removed.
	Removed []Field

	// OldOrder is the order of fields in the old list. It will be empty if the order has not changed.
	OldOrder []string

	// NewOrder is the order of fields in the new list. It will be empty if the order has not changed.
	NewOrder []string
}

func compareObjectType(oldObj, newObj ObjectType) ObjectTypeDiff {
	diff := ObjectTypeDiff{
		Name: oldObj.TypeName(),
	}

	diff.KeyFieldsDiff = compareFields(oldObj.KeyFields, newObj.KeyFields)
	diff.ValueFieldsDiff = compareFields(oldObj.ValueFields, newObj.ValueFields)

	return diff
}

func compareFields(oldFields, newFields []Field) FieldsDiff {
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
			fieldDiff := compareField(oldField, newField)
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

// Empty returns true if the object type diff has no changes.
func (o ObjectTypeDiff) Empty() bool {
	return o.KeyFieldsDiff.Empty() && o.ValueFieldsDiff.Empty()
}

func (o ObjectTypeDiff) HasCompatibleChanges() bool {
	// key fields can't be added, removed, or changed, but they can be reordered
	// value fields can't be removed or changed, but they can be added or reordered
	return len(o.KeyFieldsDiff.Added) == 0 &&
		len(o.KeyFieldsDiff.Removed) == 0 &&
		len(o.KeyFieldsDiff.Changed) == 0 &&
		len(o.ValueFieldsDiff.Removed) == 0 &&
		len(o.ValueFieldsDiff.Changed) == 0
}

// Empty returns true if the field diff has no changes.
func (d FieldsDiff) Empty() bool {
	if len(d.Added) != 0 || len(d.Changed) != 0 || len(d.Removed) != 0 {
		return false
	}

	return !d.OrderChanged()
}

// OrderChanged returns true if the field order changed.
func (d FieldsDiff) OrderChanged() bool {
	if len(d.OldOrder) == 0 && len(d.NewOrder) == 0 {
		return false
	}

	return true
}
