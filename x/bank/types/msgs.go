package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// bank message types
const (
	TypeMsgSend                = "send"
	TypeMsgMultiSend           = "multisend"
	TypeMsgUpdateDenomMetadata = "updatedenommetada"
)

var _ sdk.Msg = &MsgSend{}

// NewMsgSend - construct a msg to send coins from one account to another.
//
//nolint:interfacer
func NewMsgSend(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgSend {
	return &MsgSend{FromAddress: fromAddr.String(), ToAddress: toAddr.String(), Amount: amount}
}

// Route Implements Msg.
func (msg MsgSend) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSend) Type() string { return TypeMsgSend }

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
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

var _ sdk.Msg = &MsgMultiSend{}

// NewMsgMultiSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgMultiSend(in []Input, out []Output) *MsgMultiSend {
	return &MsgMultiSend{Inputs: in, Outputs: out}
}

// Route Implements Msg
func (msg MsgMultiSend) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMultiSend) Type() string { return TypeMsgMultiSend }

// ValidateBasic Implements Msg.
func (msg MsgMultiSend) ValidateBasic() error {
	return ValidateInputsOutputs(msg.Inputs, msg.Outputs)
}

// GetSignBytes Implements Msg.
func (msg MsgMultiSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgMultiSend) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(msg.Inputs))
	for i, in := range msg.Inputs {
		inAddr, _ := sdk.AccAddressFromBech32(in.Address)
		addrs[i] = inAddr
	}

	return addrs
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(in.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid input address: %s", err)
	}

	if !in.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}

	if !in.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}

	return nil
}

// NewInput - create a transaction input, used with MsgMultiSend
//
//nolint:interfacer
func NewInput(addr sdk.AccAddress, coins sdk.Coins) Input {
	return Input{
		Address: addr.String(),
		Coins:   coins,
	}
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(out.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid output address: %s", err)
	}

	if !out.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}

	if !out.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}

	return nil
}

// NewOutput - create a transaction output, used with MsgMultiSend
//
//nolint:interfacer
func NewOutput(addr sdk.AccAddress, coins sdk.Coins) Output {
	return Output{
		Address: addr.String(),
		Coins:   coins,
	}
}

// ValidateInputsOutputs validates that:
// - There's at least one input.
// - There's at least one output.
// - There's either exactly one input or exactly one output.
// - Each respective input and output is valid.
// - The sum of inputs equals the sum of outputs.
func ValidateInputsOutputs(inputs []Input, outputs []Output) error {
	if len(inputs) == 0 {
		return ErrNoInputs
	}
	if len(outputs) == 0 {
		return ErrNoOutputs
	}
	if len(inputs) != 1 && len(outputs) != 1 {
		return ErrManyToMany
	}

	var totalIn, totalOut sdk.Coins

	for _, input := range inputs {
		if err := input.ValidateBasic(); err != nil {
			return err
		}

		totalIn = totalIn.Add(input.Coins...)
	}

	for _, out := range outputs {
		if err := out.ValidateBasic(); err != nil {
			return err
		}

		totalOut = totalOut.Add(out.Coins...)
	}

	// The coins.IsEqual(coins) function panics if both have the same number of denoms,
	// but they're different. We don't want the panic here, so check the coins manually.
	// Assume that .Add returns sorted Coins.
	if len(totalIn) != len(totalOut) {
		return ErrInputOutputMismatch
	}
	for i, coinIn := range totalIn {
		if coinIn.Denom != totalOut[i].Denom || !coinIn.Amount.Equal(totalOut[i].Amount) {
			return ErrInputOutputMismatch
		}
	}

	return nil
}

var _ sdk.Msg = &MsgUpdateDenomMetadata{}

// NewMsgUpdateDenomMetadata - construct a message to update denom metadata
func NewMsgUpdateDenomMetadata(fromAddr, title string, description string, metadata *Metadata) *MsgUpdateDenomMetadata {
	return &MsgUpdateDenomMetadata{
		FromAddress: fromAddr,
		Title:       title,
		Description: description,
		Metadata:    metadata,
	}
}

// Route Implements Msg
func (msg MsgUpdateDenomMetadata) Route() string {
	return RouterKey
}

// Type Implements Msg
func (msg MsgUpdateDenomMetadata) Type() string {
	return TypeMsgUpdateDenomMetadata
}

// ValidateBasic Implements Msg.
func (msg MsgUpdateDenomMetadata) ValidateBasic() error {
	return msg.Metadata.Validate()
}

// GetSignBytes Implements Msg.
func (msg MsgUpdateDenomMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgUpdateDenomMetadata) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}
