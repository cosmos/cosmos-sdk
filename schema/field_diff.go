package schema

type FieldDiff struct {
	Name        string
	OldKind     Kind
	NewKind     Kind
	OldNullable bool
	NewNullable bool
	OldEnumType string
	NewEnumType string
}

func DiffField(oldField, newField Field) FieldDiff {
	diff := FieldDiff{
		Name: oldField.Name,
	}
	if oldField.Kind != newField.Kind {
		diff.OldKind = oldField.Kind
		diff.NewKind = newField.Kind
	}
	if oldField.Nullable != newField.Nullable {
		diff.OldNullable = oldField.Nullable
		diff.NewNullable = newField.Nullable
	}
	if oldField.EnumType.Name != newField.EnumType.Name {
		diff.OldEnumType = oldField.EnumType.Name
		diff.NewEnumType = newField.EnumType.Name
	}
	return diff
}

func (d FieldDiff) Empty() bool {
	return !d.KindChanged() && !d.NullableChanged() && !d.EnumTypeChanged()
}

func (d FieldDiff) KindChanged() bool {
	return d.OldKind != d.NewKind
}

func (d FieldDiff) NullableChanged() bool {
	return d.OldNullable != d.NewNullable
}

func (d FieldDiff) EnumTypeChanged() bool {
	return d.OldEnumType != d.NewEnumType
}
