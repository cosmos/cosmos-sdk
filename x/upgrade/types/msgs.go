package types

import (
	errorsmod "cosmossdk.io/errors"

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

// ValidateBasic does a sanity check on the provided data.
func (m *MsgSoftwareUpgrade) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}

	if err := m.Plan.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "plan")
	}

	return nil
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

// ValidateBasic does a sanity check on the provided data.
func (m *MsgCancelUpgrade) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}

	return nil
}

// GetSigners returns the expected signers for MsgCancelUpgrade.
func (m *MsgCancelUpgrade) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
