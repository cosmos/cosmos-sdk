package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Vetsing message types and routes
const (
	TypeMsgCreateVestingAccount         = "msg_create_vesting_account"
	TypeMsgCreateClawbackVestingAccount = "msg_create_clawback_vesting_account"
	TypeMsgClawback                     = "msg_clawback"
	TypeMsgCreatePermanentLockedAccount = "msg_create_permanent_locked_account"
	TypeMsgCreatePeriodicVestingAccount = "msg_create_periodic_vesting_account"
)

var _, _, _, _, _ sdk.Msg = &MsgCreateVestingAccount{}, &MsgCreateVestingAccount{}, &MsgClawback{}, &MsgCreatePermanentLockedAccount{}, &MsgCreatePeriodicVestingAccount{}

// NewMsgCreateVestingAccount returns a reference to a new MsgCreateVestingAccount.
//
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
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgCreateVestingAccount.
func (msg MsgCreateVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
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

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	if err := sdk.VerifyAddressFormat(to); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address: %s", err)
	}

	lockupCoins := sdk.NewCoins()
	for i, period := range msg.LockupPeriods {
		if period.Length < 1 {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
		lockupCoins = lockupCoins.Add(period.Amount...)
	}

	vestingCoins := sdk.NewCoins()
	for i, period := range msg.VestingPeriods {
		if period.Length < 1 {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
		vestingCoins = vestingCoins.Add(period.Amount...)
	}

	// If both schedules are present, the must describe the same total amount.
	// IsEqual can panic, so use (a == b) <=> (a <= b && b <= a).
	if len(msg.LockupPeriods) > 0 && len(msg.VestingPeriods) > 0 &&
		!(lockupCoins.IsAllLTE(vestingCoins) && vestingCoins.IsAllLTE(lockupCoins)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "vesting and lockup schedules must have same total coins")
	}

	return nil
}

// NewMsgClawback returns a reference to a new MsgClawback.
// The dest address may be nil - defaulting to the funder.
//
//nolint:interfacer
func NewMsgClawback(funder, addr, dest sdk.AccAddress) *MsgClawback {
	var destString string
	if dest != nil {
		destString = dest.String()
	}
	return &MsgClawback{
		FunderAddress: funder.String(),
		Address:       addr.String(),
		DestAddress:   destString,
	}
}

// Route returns the message route for a MsgClawback.
func (msg MsgClawback) Route() string { return RouterKey }

// Type returns the message type for a MsgClawback.
func (msg MsgClawback) Type() string { return TypeMsgClawback }

// GetSigners returns the expected signers for a MsgClawback.
func (msg MsgClawback) GetSigners() []sdk.AccAddress {
	funder, err := sdk.AccAddressFromBech32(msg.FunderAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{funder}
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgClawback.
func (msg MsgClawback) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&msg))
}

// ValidateBasic Implements Msg.
func (msg MsgClawback) ValidateBasic() error {
	funder, err := sdk.AccAddressFromBech32(msg.GetFunderAddress())
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(funder); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid funder address: %s", err)
	}

	addr, err := sdk.AccAddressFromBech32(msg.GetAddress())
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid account address: %s", err)
	}

	if msg.GetDestAddress() != "" {
		dest, err := sdk.AccAddressFromBech32(msg.GetDestAddress())
		if err != nil {
			return err
		}
		if err := sdk.VerifyAddressFormat(dest); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid destination address: %s", err)
		}
	}

	return nil
}

// NewMsgCreatePermanentLockedAccount returns a reference to a new MsgCreatePermanentLockedAccount.
//
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
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap(msg.Amount.String())
	}

	return nil
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{from}
}

// NewMsgCreatePeriodicVestingAccount returns a reference to a new MsgCreatePeriodicVestingAccount.
//
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
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
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
		if !period.Amount.IsValid() {
			return sdkerrors.ErrInvalidCoins.Wrap(period.Amount.String())
		}

		if !period.Amount.IsAllPositive() {
			return sdkerrors.ErrInvalidCoins.Wrap(period.Amount.String())
		}

		if period.Length < 1 {
			return fmt.Errorf("invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
	}

	return nil
}
