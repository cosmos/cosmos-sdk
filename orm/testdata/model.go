package testdata

import (
	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrTest = errors.Register("orm_testdata", 9999, "test")
)

func (g GroupMember) PrimaryKey() []byte {
	result := make([]byte, 0, len(g.Group)+len(g.Member))
	result = append(result, g.Group...)
	result = append(result, g.Member...)
	return result
}

func (g GroupInfo) PrimaryKey() []byte {
	return orm.EncodeSequence(g.GroupId)
}

func (g GroupInfo) ValidateBasic() error {
	if g.Description == "invalid" {
		return errors.Wrap(ErrTest, "description")
	}
	return nil
}

func (g GroupMember) ValidateBasic() error {
	return nil
}
