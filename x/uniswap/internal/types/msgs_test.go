package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// test ValidateBasic for MsgSwapOrder
func TestMsgSwapOrder(t *testing.T) {
	tests := []struct {
		name       string
		msg        MsgSwapOrder
		expectPass bool
	}{
		{"empty swap denomination", NewMsgSwapOrder(emptyDenom, amount, bound, deadline, sender, recipient, true), false},
		{"empty coin", NewMsgSwapOrder(denom0, sdk.NewCoins(sdk.NewCoin(denom1, zero)), bound, deadline, sender, recipient, true), false},
		{"no coin", NewMsgSwapOrder(denom0, sdk.Coins{}, bound, deadline, sender, recipient, true), false},
		{"too many coins", NewMsgSwapOrder(denom0, sdk.NewCoins(coin, sdk.NewCoin(denom2, baseValue)), bound, deadline, sender, recipient, true), false},
		{"swap and coin denomination are equal", NewMsgSwapOrder(denom0, sdk.NewCoins(sdk.NewCoin(denom0, baseValue)), bound, deadline, sender, recipient, true), false},
		{"bound is not positive", NewMsgSwapOrder(denom0, amount, zero, deadline, sender, recipient, true), false},
		{"deadline not initialized", NewMsgSwapOrder(denom0, amount, bound, emptyTime, sender, recipient, true), false},
		{"no sender", NewMsgSwapOrder(denom0, amount, bound, deadline, emptyAddr, recipient, true), false},
		{"no recipient", NewMsgSwapOrder(denom0, amount, bound, deadline, sender, emptyAddr, true), false},
		{"valid MsgSwapOrder", NewMsgSwapOrder(denom0, amount, bound, deadline, sender, recipient, true), true},
		{"sender and recipient are same", NewMsgSwapOrder(denom0, amount, bound, deadline, sender, sender, true), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectPass {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
			}
		})
	}
}

// test ValidateBasic for MsgAddLiquidity
func TestMsgAddLiquidity(t *testing.T) {
	tests := []struct {
		name       string
		msg        MsgAddLiquidity
		expectPass bool
	}{
		{"invalid withdraw amount", NewMsgAddLiquidity(denom1, zero, one, one, deadline, sender), false},
		{"empty exchange denom", NewMsgAddLiquidity(emptyDenom, baseValue, one, one, deadline, sender), false},
		{"invalid minumum liquidity bound", NewMsgAddLiquidity(denom1, baseValue, zero, one, deadline, sender), false},
		{"invalid maximum coins bound", NewMsgAddLiquidity(denom1, baseValue, one, zero, deadline, sender), false},
		{"deadline not initialized", NewMsgAddLiquidity(denom1, baseValue, one, one, emptyTime, sender), false},
		{"empty sender", NewMsgAddLiquidity(denom1, baseValue, one, one, deadline, emptyAddr), false},
		{"valid MsgAddLiquidity", NewMsgAddLiquidity(denom1, baseValue, one, one, deadline, sender), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectPass {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
			}
		})
	}
}

// test ValidateBasic for MsgRemoveLiquidity
func TestMsgRemoveLiquidity(t *testing.T) {
	tests := []struct {
		name       string
		msg        MsgRemoveLiquidity
		expectPass bool
	}{
		{"invalid deposit amount", NewMsgRemoveLiquidity(denom1, zero, one, one, deadline, sender), false},
		{"empty exchange denom", NewMsgRemoveLiquidity(emptyDenom, baseValue, one, one, deadline, sender), false},
		{"invalid minimum native bound", NewMsgRemoveLiquidity(denom1, baseValue, zero, one, deadline, sender), false},
		{"invalid minumum coins bound", NewMsgRemoveLiquidity(denom1, baseValue, one, zero, deadline, sender), false},
		{"deadline not initialized", NewMsgRemoveLiquidity(denom1, baseValue, one, one, emptyTime, sender), false},
		{"empty sender", NewMsgRemoveLiquidity(denom1, baseValue, one, one, deadline, emptyAddr), false},
		{"valid MsgRemoveLiquidity", NewMsgRemoveLiquidity(denom1, baseValue, one, one, deadline, sender), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectPass {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
			}
		})
	}

}
