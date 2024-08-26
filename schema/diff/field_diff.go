package diff

import "cosmossdk.io/schema"

// FieldDiff represents the difference between two fields.
// The KindChanged, NullableChanged, and EnumTypeChanged methods can be used to determine
// what specific changes were made to the field.
type FieldDiff struct {
	// Name is the name of the field.
	Name string

	// OldKind is the old kind of the field. It will be InvalidKind if there was no change.
	OldKind schema.Kind

	// NewKind is the new kind of the field. It will be InvalidKind if there was no change.
	NewKind schema.Kind

	// OldNullable is the old nullable property of the field.
	OldNullable bool

	// NewNullable is the new nullable property of the field.
	NewNullable bool

	// OldEnumType is the name of the old enum type of the field.
	// It will be empty if the field is not an enum type or if there was no change.
	OldEnumType string

	// NewEnumType is the name of the new enum type of the field.
	// It will be empty if the field is not an enum type or if there was no change.
	NewEnumType string
}

func compareField(oldField, newField schema.Field) FieldDiff {
	diff := FieldDiff{
		Name: oldField.Name,
	}
	if oldField.Kind != newField.Kind {
		diff.OldKind = oldField.Kind
		diff.NewKind = newField.Kind
	}

	diff.OldNullable = oldField.Nullable
	diff.NewNullable = newField.Nullable

	if oldField.EnumType.Name != newField.EnumType.Name {
		diff.OldEnumType = oldField.EnumType.Name
		diff.NewEnumType = newField.EnumType.Name
	}
	return diff
}

// Empty returns true if the field diff has no changes.
func (d FieldDiff) Empty() bool {
	return !d.KindChanged() && !d.NullableChanged() && !d.EnumTypeChanged()
}

// KindChanged returns true if the field kind changed.
func (d FieldDiff) KindChanged() bool {
	return d.OldKind != d.NewKind
}

// NullableChanged returns true if the field nullable property changed.
func (d FieldDiff) NullableChanged() bool {
	return d.OldNullable != d.NewNullable
}

// EnumTypeChanged returns true if the field enum type changed.
// Note that if the enum type name remained the same but the values of
// the enum type changed, that won't be reported here but rather in the
// ModuleSchemaDiff's ChangedEnumTypes field.
func (d FieldDiff) EnumTypeChanged() bool {
	return d.OldEnumType != d.NewEnumType
}
