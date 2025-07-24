package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Error types
var (
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1, "invalid address")
	ErrInvalidRequest = errorsmod.Register(ModuleName, 2, "invalid request")
)

// TypeMsgCreateVestingAccount defines the type value for a MsgCreateVestingAccount.
const TypeMsgCreateVestingAccount = "msg_create_vesting_account"

// TypeMsgCreatePermanentLockedAccount defines the type value for a MsgCreatePermanentLockedAccount.
const TypeMsgCreatePermanentLockedAccount = "msg_create_permanent_locked_account"

// TypeMsgCreatePeriodicVestingAccount defines the type value for a MsgCreateVestingAccount.
const TypeMsgCreatePeriodicVestingAccount = "msg_create_periodic_vesting_account"

// TypeMsgCreateClawbackVestingAccount defines the type value for a MsgCreateClawbackVestingAcount.
const TypeMsgCreateClawbackVestingAccount = "msg_create_clawback_vesting_account"

// TypeMsgClawback defines the type value for a MsgClawback.
const TypeMsgClawback = "msg_clawback"

// TypeMsgReturnGrants defines the type value for a MsgReturnGrants.
const TypeMsgReturnGrants = "msg_return_grants"

var _ sdk.Msg = &MsgCreateVestingAccount{}

var _ sdk.Msg = &MsgCreatePermanentLockedAccount{}

var _ sdk.Msg = &MsgCreatePeriodicVestingAccount{}

var _ sdk.Msg = &MsgCreateClawbackVestingAccount{}

var _ sdk.Msg = &MsgClawback{}

var _ sdk.Msg = &MsgReturnGrants{}

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

// NewMsgCreatePermanentLockedAccount returns a reference to a new MsgCreatePermanentLockedAccount.
func NewMsgCreatePermanentLockedAccount(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgCreatePermanentLockedAccount {
	return &MsgCreatePermanentLockedAccount{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
	}
}

// NewMsgCreatePeriodicVestingAccount returns a reference to a new MsgCreatePeriodicVestingAccount.
//
//nolint:interfacer
func NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr sdk.AccAddress, startTime int64, periods []Period, merge bool) *MsgCreatePeriodicVestingAccount {
	return &MsgCreatePeriodicVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		VestingPeriods: periods,
		Merge:          merge,
	}
}

// NewMsgCreateClawbackVestingAccount returns a reference to a new MsgCreateClawbackVestingAccount.
//
//nolint:interfacer
func NewMsgCreateClawbackVestingAccount(fromAddr, toAddr sdk.AccAddress, startTime int64, lockupPeriods, vestingPeriods []Period, merge bool) *MsgCreateClawbackVestingAccount {
	return &MsgCreateClawbackVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		LockupPeriods:  lockupPeriods,
		VestingPeriods: vestingPeriods,
		Merge:          merge,
	}
}

// Route returns the message route for a MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) Route() string { return RouterKey }

// Type returns the message type for a MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) Type() string { return TypeMsgCreateClawbackVestingAccount }

// GetSigners returns the expected signers for a MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// ValidateBasic Implements Msg.
func (msg MsgCreateClawbackVestingAccount) ValidateBasic() error {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(from); err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid sender address: %s", err)
	}

	if err := sdk.VerifyAddressFormat(to); err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid recipient address: %s", err)
	}

	lockupCoins := sdk.NewCoins()
	for i, period := range msg.LockupPeriods {
		if period.Length < 1 {
			return errorsmod.Wrapf(ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
		if !period.Amount.IsValid() {
			return errorsmod.Wrapf(ErrInvalidRequest, "invalid period amount in period %d: %s", i, err)
		}
		lockupCoins = lockupCoins.Add(period.Amount...)
	}

	vestingCoins := sdk.NewCoins()
	for i, period := range msg.VestingPeriods {
		if period.Length < 1 {
			return errorsmod.Wrapf(ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
		if !period.Amount.IsValid() {
			return errorsmod.Wrapf(ErrInvalidRequest, "invalid period amount in period %d: %s", i, err)
		}
		vestingCoins = vestingCoins.Add(period.Amount...)
	}

	// If both schedules are present, the must describe the same total amount.
	// IsEqual can panic, so use (a == b) <=> (a <= b && b <= a).
	if len(msg.LockupPeriods) > 0 && len(msg.VestingPeriods) > 0 &&
		!(lockupCoins.IsAllLTE(vestingCoins) && vestingCoins.IsAllLTE(lockupCoins)) {
		return errorsmod.Wrapf(ErrInvalidRequest, "vesting and lockup schedules must have same total coins")
	}

	return nil
}

// NewMsgClawback returns a reference to a new MsgClawback.
// The dest address may be nil - defaulting to the funder.
//
//nolint:interfacer
func NewMsgClawback(funder, addr, dest sdk.AccAddress) *MsgClawback {
	destString := ""
	if dest != nil {
		destString = dest.String()
	}
	return &MsgClawback{
		FunderAddress: funder.String(),
		Address:       addr.String(),
		DestAddress:   destString,
	}
}

// GetSigners returns the expected signers for a MsgClawback.
func (msg MsgClawback) GetSigners() []sdk.AccAddress {
	funder, err := sdk.AccAddressFromBech32(msg.FunderAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{funder}
}

// ValidateBasic Implements Msg.
func (msg MsgClawback) ValidateBasic() error {
	funder, err := sdk.AccAddressFromBech32(msg.GetFunderAddress())
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(funder); err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid funder address: %s", err)
	}

	addr, err := sdk.AccAddressFromBech32(msg.GetAddress())
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid account address: %s", err)
	}

	if msg.GetDestAddress() != "" {
		dest, err := sdk.AccAddressFromBech32(msg.GetDestAddress())
		if err != nil {
			return err
		}
		if err := sdk.VerifyAddressFormat(dest); err != nil {
			return errorsmod.Wrapf(ErrInvalidAddress, "invalid destination address: %s", err)
		}
	}

	return nil
}

// NewMsgReturnGrants returns a reference to a new MsgReturnGrants.
//
//nolint:interfacer
func NewMsgReturnGrants(addr sdk.AccAddress) *MsgReturnGrants {
	return &MsgReturnGrants{
		Address: addr.String(),
	}
}

// Route returns the message route for a MsgReturnGrants.
func (msg MsgReturnGrants) Route() string { return RouterKey }

// Type returns the message type for a MsgReturnGrants.
func (msg MsgReturnGrants) Type() string { return TypeMsgReturnGrants }

// GetSigners returns the expected signers for a MsgReturnGrants.
func (msg MsgReturnGrants) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic Implements Msg.
func (msg MsgReturnGrants) ValidateBasic() error {
	addr, err := sdk.AccAddressFromBech32(msg.GetAddress())
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid account address: %s", err)
	}
	return nil
}
