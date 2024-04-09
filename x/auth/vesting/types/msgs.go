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
func NewMsgCreateVestingAccount(fromAddr, toAddr string, amount sdk.Coins, endTime int64, delayed bool) *MsgCreateVestingAccount {
	return &MsgCreateVestingAccount{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      amount,
		EndTime:     endTime,
		Delayed:     delayed,
	}
}

// NewMsgCreatePermanentLockedAccount returns a reference to a new MsgCreatePermanentLockedAccount.
func NewMsgCreatePermanentLockedAccount(fromAddr, toAddr string, amount sdk.Coins) *MsgCreatePermanentLockedAccount {
	return &MsgCreatePermanentLockedAccount{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      amount,
	}
}

// NewMsgCreatePeriodicVestingAccount returns a reference to a new MsgCreatePeriodicVestingAccount.
func NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr string, startTime int64, periods []Period) *MsgCreatePeriodicVestingAccount {
	return &MsgCreatePeriodicVestingAccount{
		FromAddress:    fromAddr,
		ToAddress:      toAddr,
		StartTime:      startTime,
		VestingPeriods: periods,
	}
}
