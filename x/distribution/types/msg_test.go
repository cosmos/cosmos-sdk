package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// test ValidateBasic for MsgSetWithdrawAddress
func TestMsgSetWithdrawAddress(t *testing.T) {
	tests := []struct {
		delegatorAddr sdk.AccAddress
		withdrawAddr  sdk.AccAddress
		expectPass    bool
	}{
		{delAddr1, delAddr2, true},
		{delAddr1, delAddr1, true},
		{emptyDelAddr, delAddr1, false},
		{delAddr1, emptyDelAddr, false},
		{emptyDelAddr, emptyDelAddr, false},
	}

	for i, tc := range tests {
		msg := NewMsgSetWithdrawAddress(tc.delegatorAddr, tc.withdrawAddr)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test index: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test index: %v", i)
		}
	}
}

// test ValidateBasic for MsgWithdrawDelegatorReward
func TestMsgWithdrawDelegatorReward(t *testing.T) {
	tests := []struct {
		delegatorAddr sdk.AccAddress
		validatorAddr sdk.ValAddress
		expectPass    bool
	}{
		{delAddr1, valAddr1, true},
		{emptyDelAddr, valAddr1, false},
		{delAddr1, emptyValAddr, false},
		{emptyDelAddr, emptyValAddr, false},
	}
	for i, tc := range tests {
		msg := NewMsgWithdrawDelegatorReward(tc.delegatorAddr, tc.validatorAddr)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test index: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test index: %v", i)
		}
	}
}

// test ValidateBasic for MsgWithdrawValidatorCommission
func TestMsgWithdrawValidatorCommission(t *testing.T) {
	tests := []struct {
		validatorAddr sdk.ValAddress
		expectPass    bool
	}{
		{valAddr1, true},
		{emptyValAddr, false},
	}
	for i, tc := range tests {
		msg := NewMsgWithdrawValidatorCommission(tc.validatorAddr)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test index: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test index: %v", i)
		}
	}
}
