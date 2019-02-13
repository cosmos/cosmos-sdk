//nolint
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// name to identify transaction types
const MsgRoute = "distr"

// Verify interface at compile time
var _, _, _ sdk.Msg = &MsgSetWithdrawAddress{}, &MsgWithdrawDelegatorReward{}, &MsgWithdrawValidatorCommission{}

// msg struct for changing the withdraw address for a delegator (or validator self-delegation)
type MsgSetWithdrawAddress struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	WithdrawAddr  sdk.AccAddress `json:"withdraw_addr"`
}

func NewMsgSetWithdrawAddress(delAddr, withdrawAddr sdk.AccAddress) MsgSetWithdrawAddress {
	return MsgSetWithdrawAddress{
		DelegatorAddr: delAddr,
		WithdrawAddr:  withdrawAddr,
	}
}

func (msg MsgSetWithdrawAddress) Route() string { return MsgRoute }
func (msg MsgSetWithdrawAddress) Type() string  { return "set_withdraw_address" }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgSetWithdrawAddress) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.DelegatorAddr)}
}

// get the bytes for the message signer to sign on
func (msg MsgSetWithdrawAddress) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSetWithdrawAddress) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.WithdrawAddr.Empty() {
		return ErrNilWithdrawAddr(DefaultCodespace)
	}
	return nil
}

// msg struct for delegation withdraw from a single validator
type MsgWithdrawDelegatorReward struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
}

func NewMsgWithdrawDelegatorReward(delAddr sdk.AccAddress, valAddr sdk.ValAddress) MsgWithdrawDelegatorReward {
	return MsgWithdrawDelegatorReward{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
	}
}

func (msg MsgWithdrawDelegatorReward) Route() string { return MsgRoute }
func (msg MsgWithdrawDelegatorReward) Type() string  { return "withdraw_delegation_reward" }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawDelegatorReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.DelegatorAddr)}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawDelegatorReward) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawDelegatorReward) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	return nil
}

// msg struct for validator withdraw
type MsgWithdrawValidatorCommission struct {
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
}

func NewMsgWithdrawValidatorCommission(valAddr sdk.ValAddress) MsgWithdrawValidatorCommission {
	return MsgWithdrawValidatorCommission{
		ValidatorAddr: valAddr,
	}
}

func (msg MsgWithdrawValidatorCommission) Route() string { return MsgRoute }
func (msg MsgWithdrawValidatorCommission) Type() string  { return "withdraw_validator_rewards_all" }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawValidatorCommission) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddr.Bytes())}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawValidatorCommission) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawValidatorCommission) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	return nil
}
