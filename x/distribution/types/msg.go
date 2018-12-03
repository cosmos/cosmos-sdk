//nolint
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// name to identify transaction types
const MsgRoute = "distr"

// Verify interface at compile time
var _, _ sdk.Msg = &MsgSetWithdrawAddress{}, &MsgWithdrawDelegatorRewardsAll{}
var _, _ sdk.Msg = &MsgWithdrawDelegatorReward{}, &MsgWithdrawValidatorRewardsAll{}

//______________________________________________________________________

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
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgSetWithdrawAddress) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.WithdrawAddr == nil {
		return ErrNilWithdrawAddr(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// msg struct for delegation withdraw for all of the delegator's delegations
type MsgWithdrawDelegatorRewardsAll struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
}

func NewMsgWithdrawDelegatorRewardsAll(delAddr sdk.AccAddress) MsgWithdrawDelegatorRewardsAll {
	return MsgWithdrawDelegatorRewardsAll{
		DelegatorAddr: delAddr,
	}
}

func (msg MsgWithdrawDelegatorRewardsAll) Route() string { return MsgRoute }
func (msg MsgWithdrawDelegatorRewardsAll) Type() string  { return "withdraw_delegation_rewards_all" }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawDelegatorRewardsAll) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.DelegatorAddr)}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawDelegatorRewardsAll) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgWithdrawDelegatorRewardsAll) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

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
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgWithdrawDelegatorReward) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// msg struct for validator withdraw
type MsgWithdrawValidatorRewardsAll struct {
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
}

func NewMsgWithdrawValidatorRewardsAll(valAddr sdk.ValAddress) MsgWithdrawValidatorRewardsAll {
	return MsgWithdrawValidatorRewardsAll{
		ValidatorAddr: valAddr,
	}
}

func (msg MsgWithdrawValidatorRewardsAll) Route() string { return MsgRoute }
func (msg MsgWithdrawValidatorRewardsAll) Type() string  { return "withdraw_validator_rewards_all" }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawValidatorRewardsAll) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddr.Bytes())}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawValidatorRewardsAll) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgWithdrawValidatorRewardsAll) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	return nil
}
