package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgCreateVestingAccount{}
	_ sdk.Msg = &MsgCreatePermanentLockedAccount{}
	_ sdk.Msg = &MsgCreatePeriodicVestingAccount{}
)

// NewMsgCreateVestingAccount returns a reference to a new MsgCreateVestingAccount.
func NewMsgCreateVestingAccount(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins, endTime int64, delayed bool) *MsgCreateVestingAccount {
	return &MsgCreateVestingAccount{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
		EndTime:     endTime,
		Delayed:     delayed,
	}
}

// GetSigners returns the expected signers for a MsgCreateVestingAccount.
func (msg MsgCreateVestingAccount) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{addr}
}

// NewMsgCreatePermanentLockedAccount returns a reference to a new MsgCreatePermanentLockedAccount.
func NewMsgCreatePermanentLockedAccount(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgCreatePermanentLockedAccount {
	return &MsgCreatePermanentLockedAccount{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
	}
}

// GetSigners returns the expected signers for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{from}
}

// NewMsgCreatePeriodicVestingAccount returns a reference to a new MsgCreatePeriodicVestingAccount.
func NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr sdk.AccAddress, startTime int64, periods []Period) *MsgCreatePeriodicVestingAccount {
	return &MsgCreatePeriodicVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		VestingPeriods: periods,
	}
}

// GetSigners returns the expected signers for a MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
