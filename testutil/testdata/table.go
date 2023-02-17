package testdata

import errorsmod "cosmossdk.io/errors"

var ErrTest = errorsmod.Register("table_testdata", 2, "test")

func (g TableModel) PrimaryKeyFields() []interface{} {
	return []interface{}{g.Id}
}

func (g TableModel) ValidateBasic() error {
	if g.Name == "" {
		return errorsmod.Wrap(ErrTest, "name")
	}
	return nil
}
