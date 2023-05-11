package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_, _ sdk.Msg = &MsgSoftwareUpgrade{}, &MsgCancelUpgrade{}
)

// GetSigners returns the expected signers for MsgSoftwareUpgrade.
func (m *MsgSoftwareUpgrade) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// GetSigners returns the expected signers for MsgCancelUpgrade.
func (m *MsgCancelUpgrade) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
