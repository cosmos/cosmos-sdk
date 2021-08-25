package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TypeMsgCreateVestingAccount defines the type value for a MsgCreateVestingAccount.
const TypeMsgCreateVestingAccount = "msg_create_vesting_account"

// TypeMsgCreatePeriodicVestingAccount defines the type value for a MsgCreateVestingAccount.
const TypeMsgCreatePeriodicVestingAccount = "msg_create_periodic_vesting_account"

var _ sdk.Msg = &MsgCreateVestingAccount{}

var _ sdk.Msg = &MsgCreatePeriodicVestingAccount{}

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
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid 'from' address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
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
	addr, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{addr}
}

// NewMsgCreatePeriodicVestingAccount returns a reference to a new MsgCreatePeriodicVestingAccount.
//nolint:interfacer
func NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr sdk.AccAddress, startTime int64, periods []Period) *MsgCreatePeriodicVestingAccount {
	return &MsgCreatePeriodicVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		VestingPeriods: periods,
	}
}

// Route returns the message route for a MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) Route() string { return RouterKey }

// Type returns the message type for a MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) Type() string { return TypeMsgCreatePeriodicVestingAccount }

// GetSigners returns the expected signers for a MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&msg))
}

// ValidateBasic Implements Msg.
func (msg MsgCreatePeriodicVestingAccount) ValidateBasic() error {
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

	if msg.StartTime < 1 {
		return fmt.Errorf("invalid start time of %d, length must be greater than 0", msg.StartTime)
	}

	for i, period := range msg.VestingPeriods {
		if period.Length < 1 {
			return fmt.Errorf("invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
	}

	return nil
}
