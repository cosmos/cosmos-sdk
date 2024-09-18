package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/feemarket/types"
)

func TestMsgParams(t *testing.T) {
	t.Run("should reject a message with an invalid authority address", func(t *testing.T) {
		msg := types.NewMsgParams("invalid", types.DefaultParams())
		err := msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("should accept an empty message with a valid authority address", func(t *testing.T) {
		msg := types.NewMsgParams(sdk.AccAddress("test").String(), types.DefaultParams())
		err := msg.ValidateBasic()
		require.NoError(t, err)
	})
}
