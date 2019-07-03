package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSupply(t *testing.T) {
	initialPower := int64(100)
	initTokens := sdk.TokensFromConsensusPower(initialPower)
	nAccs := int64(4)

	ctx, _, keeper := createTestInput(t, false, initialPower, nAccs)

	total := keeper.GetSupply(ctx).Total
	expectedTotal := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(nAccs)))

	require.Equal(t, expectedTotal, total)
}
