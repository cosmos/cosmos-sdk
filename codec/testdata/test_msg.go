package testdata

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewTestMsg(addrs ...sdk.AccAddress) *TestMsg {
	return &TestMsg{
		Signers: addrs,
	}
}

var _ sdk.Msg = (*TestMsg)(nil)

func (msg *TestMsg) Route() string { return "TestMsg" }
func (msg *TestMsg) Type() string  { return "Test message" }
func (msg *TestMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg.Signers)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}
func (msg *TestMsg) ValidateBasic() error { return nil }
