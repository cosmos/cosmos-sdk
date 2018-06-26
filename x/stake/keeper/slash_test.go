package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// tests Revoke, Unrevoke
func TestRevocation(t *testing.T) {

	// setup
	ctx, _, keeper := CreateTestInput(t, false, 10)
	amt := int64(10)
	addr := addrVals[0]
	pk := PKs[0]
	pool := keeper.GetPool(ctx)
	validator := types.NewValidator(addr, pk, types.Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, amt)
	keeper.SetPool(ctx, pool)
	keeper.UpdateValidator(ctx, validator)
	keeper.SetValidatorByPubKeyIndex(ctx, validator)

	// initial state
	val, found := keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.False(t, val.GetRevoked())

	// test revoke
	keeper.Revoke(ctx, pk)
	val, found = keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.True(t, val.GetRevoked())

	// test unrevoke
	keeper.Unrevoke(ctx, pk)
	val, found = keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.False(t, val.GetRevoked())

}
