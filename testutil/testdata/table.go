package testdata

import "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrTest = errors.Register("table_testdata", 2, "test")
)

func (g TableModel) ValidateBasic() error {
	if g.Name == "" {
		return errors.Wrap(ErrTest, "name")
	}
	return nil
}
