package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TypeMsgCreateVestingAccount defines the type value for a MsgCreateVestingAccount.
const TypeMsgCreateVestingAccount = "msg_create_vesting_account"

// TypeMsgCreatePermanentLockedAccount defines the type value for a MsgCreatePermanentLockedAccount.
const TypeMsgCreatePermanentLockedAccount = "msg_create_permanent_locked_account"

var _ sdk.Msg = &MsgCreateVestingAccount{}

var _ sdk.Msg = &MsgCreatePermanentLockedAccount{}

// NewMsgCreateVestingAccount returns a reference to a new MsgCreateVestingAccount.
//nolint:interfacer
func NewMsgCreateVestingAccount(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins, endTime int64, delayed bool) *MsgCreateVestingAccount {
	return &MsgCreateVestingAccount{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
		EndTime:     endTime,
		Delayed:     delayed,
	}
}

// Route returns the message route for a MsgCreateVestingAccount.
func (msg MsgCreateVestingAccount) Route() string { return RouterKey }

// Type returns the message type for a MsgCreateVestingAccount.
func (msg MsgCreateVestingAccount) Type() string { return TypeMsgCreateVestingAccount }

// ValidateBasic Implements Msg.
func (msg MsgCreateVestingAccount) ValidateBasic() error {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(from); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	if err := sdk.VerifyAddressFormat(to); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if msg.EndTime <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid end time")
	}

	return nil
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreateVestingAccount.
func (msg MsgCreateVestingAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgCreateVestingAccount.
func (msg MsgCreateVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// NewMsgCreatePermanentLockedAccount returns a reference to a new MsgCreatePermanentLockedAccount.
//nolint:interfacer
func NewMsgCreatePermanentLockedAccount(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgCreatePermanentLockedAccount {
	return &MsgCreatePermanentLockedAccount{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
	}
}

// Route returns the message route for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) Route() string { return RouterKey }

// Type returns the message type for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) Type() string { return TypeMsgCreatePermanentLockedAccount }

// ValidateBasic Implements Msg.
func (msg MsgCreatePermanentLockedAccount) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{from}
}
