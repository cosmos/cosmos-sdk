package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Verify interface at compile time
var (
	_ sdk.Msg = (*MsgSetWithdrawAddress)(nil)
	_ sdk.Msg = (*MsgWithdrawDelegatorReward)(nil)
	_ sdk.Msg = (*MsgWithdrawValidatorCommission)(nil)
	_ sdk.Msg = (*MsgUpdateParams)(nil)
	_ sdk.Msg = (*MsgCommunityPoolSpend)(nil)
	_ sdk.Msg = (*MsgDepositValidatorRewardsPool)(nil)
)

func NewMsgSetWithdrawAddress(delAddr, withdrawAddr sdk.AccAddress) *MsgSetWithdrawAddress {
	return &MsgSetWithdrawAddress{
		DelegatorAddress: delAddr.String(),
		WithdrawAddress:  withdrawAddr.String(),
	}
}

func NewMsgWithdrawDelegatorReward(delAddr sdk.AccAddress, valAddr sdk.ValAddress) *MsgWithdrawDelegatorReward {
	return &MsgWithdrawDelegatorReward{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
	}
}

func NewMsgWithdrawValidatorCommission(valAddr sdk.ValAddress) *MsgWithdrawValidatorCommission {
	return &MsgWithdrawValidatorCommission{
		ValidatorAddress: valAddr.String(),
	}
}

// NewMsgFundCommunityPool returns a new MsgFundCommunityPool with a sender and
// a funding amount.
func NewMsgFundCommunityPool(amount sdk.Coins, depositor sdk.AccAddress) *MsgFundCommunityPool {
	return &MsgFundCommunityPool{
		Amount:    amount,
		Depositor: depositor.String(),
	}
}

// NewMsgDepositValidatorRewardsPool returns a new MsgDepositValidatorRewardsPool
// with a depositor and a funding amount.
func NewMsgDepositValidatorRewardsPool(depositor sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coins) *MsgDepositValidatorRewardsPool {
	return &MsgDepositValidatorRewardsPool{
		Amount:           amount,
		Depositor:        depositor.String(),
		ValidatorAddress: valAddr.String(),
	}
}
