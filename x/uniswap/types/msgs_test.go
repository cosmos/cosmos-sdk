package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// test ValidateBasic for MsgSwapOrder
func TestMsgSwapOrder(t *testing.T) {
	tests := []struct {
		newCoinDenom string
		signer       sdk.AccAddress
		expectPass   bool
	}{
		{emptyDenom, addr, false},
		{denom, emptyAddr, false},
		{denom, addr, true},
	}

	for i, tc := range tests {
		msg := NewMsgCreateExchange(tc.newCoinDenom, tc.signer)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test index: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test index: %v", i)
		}
	}
}
