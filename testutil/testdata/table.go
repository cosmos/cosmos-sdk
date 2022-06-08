package testdata

import (
	sdkerrors "cosmossdk.io/errors"
)

var ErrTest = sdkerrors.Register("table_testdata", 2, "test")

func (g TableModel) PrimaryKeyFields() []interface{} {
	return []interface{}{g.Id}
}

func (g TableModel) ValidateBasic() error {
	if g.Name == "" {
		return sdkerrors.Wrap(ErrTest, "name")
	}
	return nil
}
