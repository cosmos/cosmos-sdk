package testdata

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var ErrTest = errors.Register("table_testdata", 2, "test")

func (g TableModel) PrimaryKeyFields() []interface{} {
	return []interface{}{g.Id}
}

func (g TableModel) ValidateBasic() error {
	if g.Name == "" {
		return errorsmod.Wrap(ErrTest, "name")
	}
	return nil
}
