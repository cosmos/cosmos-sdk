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
		{"no input coin", NewMsgSwapOrder(sdk.Coin{}, output, deadline, sender, recipient, true), false},
		{"zero input coin", NewMsgSwapOrder(sdk.NewCoin(denom0, sdk.ZeroInt()), output, deadline, sender, recipient, true), false},
		{"no output coin", NewMsgSwapOrder(input, sdk.Coin{}, deadline, sender, recipient, false), false},
		{"zero output coin", NewMsgSwapOrder(input, sdk.NewCoin(denom1, sdk.ZeroInt()), deadline, sender, recipient, true), false},
		{"swap and coin denomination are equal", NewMsgSwapOrder(input, sdk.NewCoin(denom0, amt), deadline, sender, recipient, true), false},
		{"deadline not initialized", NewMsgSwapOrder(input, output, emptyTime, sender, recipient, true), false},
		{"no sender", NewMsgSwapOrder(input, output, deadline, emptyAddr, recipient, true), false},
		{"no recipient", NewMsgSwapOrder(input, output, deadline, sender, emptyAddr, true), false},
		{"valid MsgSwapOrder", NewMsgSwapOrder(input, output, deadline, sender, recipient, true), true},
		{"sender and recipient are same", NewMsgSwapOrder(input, output, deadline, sender, sender, true), true},
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
		{"no deposit coin", NewMsgAddLiquidity(sdk.Coin{}, amt, sdk.OneInt(), deadline, sender), false},
		{"zero deposit coin", NewMsgAddLiquidity(sdk.NewCoin(denom1, sdk.ZeroInt()), amt, sdk.OneInt(), deadline, sender), false},
		{"invalid withdraw amount", NewMsgAddLiquidity(input, sdk.ZeroInt(), sdk.OneInt(), deadline, sender), false},
		{"invalid minumum reward bound", NewMsgAddLiquidity(input, amt, sdk.ZeroInt(), deadline, sender), false},
		{"deadline not initialized", NewMsgAddLiquidity(input, amt, sdk.OneInt(), emptyTime, sender), false},
		{"empty sender", NewMsgAddLiquidity(input, amt, sdk.OneInt(), deadline, emptyAddr), false},
		{"valid MsgAddLiquidity", NewMsgAddLiquidity(input, amt, sdk.OneInt(), deadline, sender), true},
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
		{"no withdraw coin", NewMsgRemoveLiquidity(sdk.Coin{}, amt, sdk.OneInt(), deadline, sender), false},
		{"zero withdraw coin", NewMsgRemoveLiquidity(sdk.NewCoin(denom1, sdk.ZeroInt()), amt, sdk.OneInt(), deadline, sender), false},
		{"invalid deposit amount", NewMsgRemoveLiquidity(input, sdk.ZeroInt(), sdk.OneInt(), deadline, sender), false},
		{"invalid minimum native bound", NewMsgRemoveLiquidity(input, amt, sdk.ZeroInt(), deadline, sender), false},
		{"deadline not initialized", NewMsgRemoveLiquidity(input, amt, sdk.OneInt(), emptyTime, sender), false},
		{"empty sender", NewMsgRemoveLiquidity(input, amt, sdk.OneInt(), deadline, emptyAddr), false},
		{"valid MsgRemoveLiquidity", NewMsgRemoveLiquidity(input, amt, sdk.OneInt(), deadline, sender), true},
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
