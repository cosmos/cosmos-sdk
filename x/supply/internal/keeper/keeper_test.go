package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

func TestSupply(t *testing.T) {
	initialPower := int64(100)
	initTokens := sdk.TokensFromConsensusPower(initialPower)
	nAccs := int64(4)

	ctx, _, keeper := createTestInput(t, false, initialPower, nAccs)

	total := keeper.GetSupply(ctx).GetTotal()
	expectedTotal := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(nAccs)))

	require.Equal(t, expectedTotal, total)
}

func TestValidatePermissions(t *testing.T) {
	nAccs := int64(0)
	initialPower := int64(100)
	_, _, keeper := createTestInput(t, false, initialPower, nAccs)

	err := keeper.ValidatePermissions(multiPermAcc)
	require.NoError(t, err)

	err = keeper.ValidatePermissions(randomPermAcc)
	require.NoError(t, err)

	// unregistered permissions
	otherAcc := types.NewEmptyModuleAccount("other", "other")
	err = keeper.ValidatePermissions(otherAcc)
	require.Error(t, err)
}
