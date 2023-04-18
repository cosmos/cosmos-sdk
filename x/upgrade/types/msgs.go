package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	_, _ sdk.Msg            = &MsgSoftwareUpgrade{}, &MsgCancelUpgrade{}
	_, _ legacytx.LegacyMsg = &MsgSoftwareUpgrade{}, &MsgCancelUpgrade{}
)

// GetSignBytes implements the LegacyMsg interface.
func (m MsgSoftwareUpgrade) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for MsgSoftwareUpgrade.
func (m *MsgSoftwareUpgrade) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgCancelUpgrade) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for MsgCancelUpgrade.
func (m *MsgCancelUpgrade) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
