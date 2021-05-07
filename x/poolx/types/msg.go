package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Return address that must sign over msg.GetSignBytes()
func (msg MsgCreatePool) GetSigners() []sdk.AccAddress {
	return nil // TODO(levi) understand & implement this
}

// quick validity check
func (msg MsgCreatePool) ValidateBasic() error {
	return nil // TODO(levi) understand & implement this
}

// Return address that must sign over msg.GetSignBytes()
func (msg MsgFundPool) GetSigners() []sdk.AccAddress {
	return nil // TODO(levi) understand & implement this
}

// quick validity check
func (msg MsgFundPool) ValidateBasic() error {
	return nil // TODO(levi) understand & implement this
}
