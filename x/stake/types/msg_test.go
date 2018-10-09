package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

var (
	coinPos  = sdk.NewInt64Coin("steak", 1000)
	coinZero = sdk.NewInt64Coin("steak", 0)
	coinNeg  = sdk.NewInt64Coin("steak", -10000)
)

// test ValidateBasic for MsgCreateValidator
func TestMsgCreateValidator(t *testing.T) {
	commission1 := NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	commission2 := NewCommissionMsg(sdk.NewDec(5), sdk.NewDec(5), sdk.NewDec(5))

	tests := []struct {
		name, moniker, identity, website, details string
		commissionMsg                             CommissionMsg
		validatorAddr                             sdk.ValAddress
		pubkey                                    crypto.PubKey
		bond                                      sdk.Coin
		expectPass                                bool
	}{
		{"basic good", "a", "b", "c", "d", commission1, addr1, pk1, coinPos, true},
		{"partial description", "", "", "c", "", commission1, addr1, pk1, coinPos, true},
		{"empty description", "", "", "", "", commission2, addr1, pk1, coinPos, false},
		{"empty address", "a", "b", "c", "d", commission2, emptyAddr, pk1, coinPos, false},
		{"empty pubkey", "a", "b", "c", "d", commission1, addr1, emptyPubkey, coinPos, true},
		{"empty bond", "a", "b", "c", "d", commission2, addr1, pk1, coinZero, false},
		{"negative bond", "a", "b", "c", "d", commission2, addr1, pk1, coinNeg, false},
		{"negative bond", "a", "b", "c", "d", commission1, addr1, pk1, coinNeg, false},
	}

	for _, tc := range tests {
		description := NewDescription(tc.moniker, tc.identity, tc.website, tc.details)
		msg := NewMsgCreateValidator(tc.validatorAddr, tc.pubkey, tc.bond, description, tc.commissionMsg)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgEditValidator
func TestMsgEditValidator(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, details string
		validatorAddr                             sdk.ValAddress
		expectPass                                bool
	}{
		{"basic good", "a", "b", "c", "d", addr1, true},
		{"partial description", "", "", "c", "", addr1, true},
		{"empty description", "", "", "", "", addr1, false},
		{"empty address", "a", "b", "c", "d", emptyAddr, false},
	}

	for _, tc := range tests {
		description := NewDescription(tc.moniker, tc.identity, tc.website, tc.details)
		newRate := sdk.ZeroDec()

		msg := NewMsgEditValidator(tc.validatorAddr, description, &newRate)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic and GetSigners for MsgCreateValidatorOnBehalfOf
func TestMsgCreateValidatorOnBehalfOf(t *testing.T) {
	commission1 := NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	commission2 := NewCommissionMsg(sdk.NewDec(5), sdk.NewDec(5), sdk.NewDec(5))

	tests := []struct {
		name, moniker, identity, website, details string
		commissionMsg                             CommissionMsg
		delegatorAddr                             sdk.AccAddress
		validatorAddr                             sdk.ValAddress
		validatorPubKey                           crypto.PubKey
		bond                                      sdk.Coin
		expectPass                                bool
	}{
		{"basic good", "a", "b", "c", "d", commission2, sdk.AccAddress(addr1), addr2, pk2, coinPos, true},
		{"partial description", "", "", "c", "", commission2, sdk.AccAddress(addr1), addr2, pk2, coinPos, true},
		{"empty description", "", "", "", "", commission1, sdk.AccAddress(addr1), addr2, pk2, coinPos, false},
		{"empty delegator address", "a", "b", "c", "d", commission1, sdk.AccAddress(emptyAddr), addr2, pk2, coinPos, false},
		{"empty validator address", "a", "b", "c", "d", commission2, sdk.AccAddress(addr1), emptyAddr, pk2, coinPos, false},
		{"empty pubkey", "a", "b", "c", "d", commission1, sdk.AccAddress(addr1), addr2, emptyPubkey, coinPos, true},
		{"empty bond", "a", "b", "c", "d", commission2, sdk.AccAddress(addr1), addr2, pk2, coinZero, false},
		{"negative bond", "a", "b", "c", "d", commission1, sdk.AccAddress(addr1), addr2, pk2, coinNeg, false},
		{"negative bond", "a", "b", "c", "d", commission2, sdk.AccAddress(addr1), addr2, pk2, coinNeg, false},
	}

	for _, tc := range tests {
		description := NewDescription(tc.moniker, tc.identity, tc.website, tc.details)
		msg := NewMsgCreateValidatorOnBehalfOf(
			tc.delegatorAddr, tc.validatorAddr, tc.validatorPubKey, tc.bond, description, tc.commissionMsg,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}

	msg := NewMsgCreateValidator(addr1, pk1, coinPos, Description{}, CommissionMsg{})
	addrs := msg.GetSigners()
	require.Equal(t, []sdk.AccAddress{sdk.AccAddress(addr1)}, addrs, "Signers on default msg is wrong")

	msg = NewMsgCreateValidatorOnBehalfOf(sdk.AccAddress(addr2), addr1, pk1, coinPos, Description{}, CommissionMsg{})
	addrs = msg.GetSigners()
	require.Equal(t, []sdk.AccAddress{sdk.AccAddress(addr2), sdk.AccAddress(addr1)}, addrs, "Signers for onbehalfof msg is wrong")
}

// test ValidateBasic for MsgDelegate
func TestMsgDelegate(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.AccAddress
		validatorAddr sdk.ValAddress
		bond          sdk.Coin
		expectPass    bool
	}{
		{"basic good", sdk.AccAddress(addr1), addr2, coinPos, true},
		{"self bond", sdk.AccAddress(addr1), addr1, coinPos, true},
		{"empty delegator", sdk.AccAddress(emptyAddr), addr1, coinPos, false},
		{"empty validator", sdk.AccAddress(addr1), emptyAddr, coinPos, false},
		{"empty bond", sdk.AccAddress(addr1), addr2, coinZero, false},
		{"negative bond", sdk.AccAddress(addr1), addr2, coinNeg, false},
	}

	for _, tc := range tests {
		msg := NewMsgDelegate(tc.delegatorAddr, tc.validatorAddr, tc.bond)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgBeginRedelegate(t *testing.T) {
	tests := []struct {
		name             string
		delegatorAddr    sdk.AccAddress
		validatorSrcAddr sdk.ValAddress
		validatorDstAddr sdk.ValAddress
		sharesAmount     sdk.Dec
		expectPass       bool
	}{
		{"regular", sdk.AccAddress(addr1), addr2, addr3, sdk.NewDecWithPrec(1, 1), true},
		{"negative decimal", sdk.AccAddress(addr1), addr2, addr3, sdk.NewDecWithPrec(-1, 1), false},
		{"zero amount", sdk.AccAddress(addr1), addr2, addr3, sdk.ZeroDec(), false},
		{"empty delegator", sdk.AccAddress(emptyAddr), addr1, addr3, sdk.NewDecWithPrec(1, 1), false},
		{"empty source validator", sdk.AccAddress(addr1), emptyAddr, addr3, sdk.NewDecWithPrec(1, 1), false},
		{"empty destination validator", sdk.AccAddress(addr1), addr2, emptyAddr, sdk.NewDecWithPrec(1, 1), false},
	}

	for _, tc := range tests {
		msg := NewMsgBeginRedelegate(tc.delegatorAddr, tc.validatorSrcAddr, tc.validatorDstAddr, tc.sharesAmount)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgBeginUnbonding(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.AccAddress
		validatorAddr sdk.ValAddress
		sharesAmount  sdk.Dec
		expectPass    bool
	}{
		{"regular", sdk.AccAddress(addr1), addr2, sdk.NewDecWithPrec(1, 1), true},
		{"negative decimal", sdk.AccAddress(addr1), addr2, sdk.NewDecWithPrec(-1, 1), false},
		{"zero amount", sdk.AccAddress(addr1), addr2, sdk.ZeroDec(), false},
		{"empty delegator", sdk.AccAddress(emptyAddr), addr1, sdk.NewDecWithPrec(1, 1), false},
		{"empty validator", sdk.AccAddress(addr1), emptyAddr, sdk.NewDecWithPrec(1, 1), false},
	}

	for _, tc := range tests {
		msg := NewMsgBeginUnbonding(tc.delegatorAddr, tc.validatorAddr, tc.sharesAmount)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}
