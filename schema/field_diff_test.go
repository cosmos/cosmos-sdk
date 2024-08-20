package schema

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_compareField(t *testing.T) {
	tests := []struct {
		oldField Field
		newField Field
		wantDiff FieldDiff
		trueF    func(FieldDiff) bool
	}{
		{
			oldField: Field{Kind: Int32Kind},
			newField: Field{Kind: Int32Kind},
			wantDiff: FieldDiff{},
			trueF:    FieldDiff.Empty,
		},
		{
			oldField: Field{Kind: StringKind},
			newField: Field{Kind: Int32Kind},
			wantDiff: FieldDiff{
				OldKind: StringKind,
				NewKind: Int32Kind,
			},
			trueF: FieldDiff.KindChanged,
		},
		{
			oldField: Field{Kind: StringKind},
			newField: Field{Kind: StringKind, Nullable: true},
			wantDiff: FieldDiff{
				NewNullable: true,
			},
			trueF: FieldDiff.NullableChanged,
		},
		{
			oldField: Field{Kind: EnumKind, EnumType: EnumType{Name: "old"}},
			newField: Field{Kind: EnumKind, EnumType: EnumType{Name: "new"}},
			wantDiff: FieldDiff{
				OldEnumType: "old",
				NewEnumType: "new",
			},
			trueF: FieldDiff.EnumTypeChanged,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			gotDiff := compareField(tt.oldField, tt.newField)
			if !reflect.DeepEqual(gotDiff, tt.wantDiff) {
				t.Errorf("compareField() = %v, want %v", gotDiff, tt.wantDiff)
			}
			if tt.trueF != nil && !tt.trueF(gotDiff) {
				t.Errorf("trueF() = false, want true")
			}
		})
	}
}
