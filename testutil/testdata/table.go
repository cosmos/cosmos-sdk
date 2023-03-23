package testdata

import "cosmossdk.io/errors"

var ErrTest = errors.Register("table_testdata", 2, "test")

func (g TableModel) PrimaryKeyFields() []any {
	return []any{g.Id}
}

func (g TableModel) ValidateBasic() error {
	if g.Name == "" {
		return errors.Wrap(ErrTest, "name")
	}
	return nil
}
