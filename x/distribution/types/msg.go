package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// Verify interface at compile time
var (
	_ sdk.Msg = (*MsgSetWithdrawAddress)(nil)
	_ sdk.Msg = (*MsgWithdrawDelegatorReward)(nil)
	_ sdk.Msg = (*MsgWithdrawValidatorCommission)(nil)
	_ sdk.Msg = (*MsgUpdateParams)(nil)
	_ sdk.Msg = (*MsgCommunityPoolSpend)(nil)
	_ sdk.Msg = (*MsgDepositValidatorRewardsPool)(nil)

	_ legacytx.LegacyMsg = (*MsgSetWithdrawAddress)(nil)
	_ legacytx.LegacyMsg = (*MsgWithdrawDelegatorReward)(nil)
	_ legacytx.LegacyMsg = (*MsgWithdrawValidatorCommission)(nil)
	_ legacytx.LegacyMsg = (*MsgUpdateParams)(nil)
	_ legacytx.LegacyMsg = (*MsgCommunityPoolSpend)(nil)
	_ legacytx.LegacyMsg = (*MsgDepositValidatorRewardsPool)(nil)
)

func NewMsgSetWithdrawAddress(delAddr, withdrawAddr sdk.AccAddress) *MsgSetWithdrawAddress {
	return &MsgSetWithdrawAddress{
		DelegatorAddress: delAddr.String(),
		WithdrawAddress:  withdrawAddr.String(),
	}
}

// Return address that must sign over msg.GetSignBytes()
func (msg MsgSetWithdrawAddress) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// get the bytes for the message signer to sign on
func (msg MsgSetWithdrawAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgWithdrawDelegatorReward(delAddr sdk.AccAddress, valAddr sdk.ValAddress) *MsgWithdrawDelegatorReward {
	return &MsgWithdrawDelegatorReward{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
	}
}

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawDelegatorReward) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawDelegatorReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

func NewMsgWithdrawValidatorCommission(valAddr sdk.ValAddress) *MsgWithdrawValidatorCommission {
	return &MsgWithdrawValidatorCommission{
		ValidatorAddress: valAddr.String(),
	}
}

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawValidatorCommission) GetSigners() []sdk.AccAddress {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	return []sdk.AccAddress{sdk.AccAddress(valAddr)}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawValidatorCommission) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// NewMsgFundCommunityPool returns a new MsgFundCommunityPool with a sender and
// a funding amount.
func NewMsgFundCommunityPool(amount sdk.Coins, depositor sdk.AccAddress) *MsgFundCommunityPool {
	return &MsgFundCommunityPool{
		Amount:    amount,
		Depositor: depositor.String(),
	}
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgFundCommunityPool) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// GetSignBytes returns the raw bytes for a MsgFundCommunityPool message that
// the expected signer needs to sign.
func (msg MsgFundCommunityPool) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes, which is the authority.
func (msg MsgCommunityPoolSpend) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the raw bytes for a MsgCommunityPoolSpend message that
// the expected signer needs to sign.
func (msg MsgCommunityPoolSpend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
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

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes, which is the depositor.
func (msg MsgDepositValidatorRewardsPool) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// GetSignBytes returns the raw bytes for a MsgDepositValidatorRewardsPool message
// that the expected signer needs to sign.
func (msg MsgDepositValidatorRewardsPool) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}
