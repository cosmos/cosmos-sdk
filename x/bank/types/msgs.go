package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	_ sdk.Msg = &MsgSend{}
	_ sdk.Msg = &MsgMultiSend{}
	_ sdk.Msg = &MsgUpdateParams{}

	_ legacytx.LegacyMsg = &MsgSend{}
	_ legacytx.LegacyMsg = &MsgMultiSend{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
)

// NewMsgSend - construct a msg to send coins from one account to another.
func NewMsgSend(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgSend {
	return &MsgSend{FromAddress: fromAddr.String(), ToAddress: toAddr.String(), Amount: amount}
}

// GetSignBytes Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

// NewMsgMultiSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgMultiSend(in Input, out []Output) *MsgMultiSend {
	return &MsgMultiSend{Inputs: []Input{in}, Outputs: out}
}

// GetSignBytes Implements Msg.
func (msg MsgMultiSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgMultiSend) GetSigners() []sdk.AccAddress {
	// should not happen as ValidateBasic would have failed
	if len(msg.Inputs) == 0 {
		return nil
	}

	addrs, _ := sdk.AccAddressFromBech32(msg.Inputs[0].Address)
	return []sdk.AccAddress{addrs}
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// NewMsgSetSendEnabled Construct a message to set one or more SendEnabled entries.
func NewMsgSetSendEnabled(authority string, sendEnabled []*SendEnabled, useDefaultFor []string) *MsgSetSendEnabled {
	return &MsgSetSendEnabled{
		Authority:     authority,
		SendEnabled:   sendEnabled,
		UseDefaultFor: useDefaultFor,
	}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgSetSendEnabled) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for MsgSoftwareUpgrade.
func (msg MsgSetSendEnabled) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}
