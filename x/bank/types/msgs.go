package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgSend{}
	_ sdk.Msg = &MsgMultiSend{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgSend - construct a msg to send coins from one account to another.
func NewMsgSend(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgSend {
	return &MsgSend{FromAddress: fromAddr.String(), ToAddress: toAddr.String(), Amount: amount}
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

// NewMsgSetSendEnabled Construct a message to set one or more SendEnabled entries.
func NewMsgSetSendEnabled(authority string, sendEnabled []*SendEnabled, useDefaultFor []string) *MsgSetSendEnabled {
	return &MsgSetSendEnabled{
		Authority:     authority,
		SendEnabled:   sendEnabled,
		UseDefaultFor: useDefaultFor,
	}
}

// GetSigners returns the expected signers for MsgSoftwareUpgrade.
func (msg MsgSetSendEnabled) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}
