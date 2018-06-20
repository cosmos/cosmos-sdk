package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondedRatio(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))
	pool := keeper.GetPool(ctx)
	pool.LooseUnbondedTokens = sdk.NewInt(1)
	pool.BondedTokens = sdk.NewInt(2)

	// bonded pool / total supply
	require.Equal(t, pool.bondedRatio(), sdk.NewRat(2).Quo(sdk.NewRat(3)))

	// avoids divide-by-zero
	pool.LooseUnbondedTokens = sdk.NewInt(0)
	pool.BondedTokens = sdk.NewInt(0)
	require.Equal(t, pool.bondedRatio(), sdk.ZeroRat())
}

func TestBondedShareExRate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))
	pool := keeper.GetPool(ctx)
	pool.BondedTokens = sdk.NewInt(3)
	pool.BondedShares = sdk.NewRat(10)

	// bonded pool / bonded shares
	require.Equal(t, pool.bondedShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.BondedShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.bondedShareExRate(), sdk.OneRat())
}

func TestUnbondingShareExRate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))
	pool := keeper.GetPool(ctx)
	pool.UnbondingTokens = sdk.NewInt(3)
	pool.UnbondingShares = sdk.NewRat(10)

	// unbonding pool / unbonding shares
	require.Equal(t, pool.unbondingShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.UnbondingShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.unbondingShareExRate(), sdk.OneRat())
}

func TestUnbondedShareExRate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))
	pool := keeper.GetPool(ctx)
	pool.UnbondedTokens = sdk.NewInt(3)
	pool.UnbondedShares = sdk.NewRat(10)

	// unbonded pool / unbonded shares
	require.Equal(t, pool.unbondedShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.UnbondedShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.unbondedShareExRate(), sdk.OneRat())
}

func TestAddTokensBonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	poolB, sharesB := poolA.addTokensBonded(sdk.NewInt(10))
	assert.Equal(t, poolB.bondedShareExRate(), sdk.OneRat())

	// correct changes to bonded shares and bonded pool
	assert.Equal(t, poolB.BondedShares, poolA.BondedShares.Add(sharesB.Amount))
	assert.Equal(t, poolB.BondedTokens, poolA.BondedTokens.Add(sdk.NewInt(10)))

	// same number of bonded shares / tokens when exchange rate is one
	assert.True(t, poolB.BondedShares.Equal(sdk.NewRatFromInt(poolB.BondedTokens)))
}

func TestRemoveSharesBonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	poolB, tokensB := poolA.removeSharesBonded(sdk.NewRat(10))
	assert.Equal(t, poolB.bondedShareExRate(), sdk.OneRat())

	// correct changes to bonded shares and bonded pool
	assert.Equal(t, poolB.BondedShares, poolA.BondedShares.Sub(sdk.NewRat(10)))
	assert.Equal(t, poolB.BondedTokens, poolA.BondedTokens.Sub(tokensB))

	// same number of bonded shares / tokens when exchange rate is one
	assert.True(t, poolB.BondedShares.Equal(sdk.NewRatFromInt(poolB.BondedTokens)))
}

func TestAddTokensUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	poolB, sharesB := poolA.addTokensUnbonded(sdk.NewInt(10))
	assert.Equal(t, poolB.unbondedShareExRate(), sdk.OneRat())

	// correct changes to unbonded shares and unbonded pool
	assert.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Add(sharesB.Amount))
	assert.Equal(t, poolB.UnbondedTokens, poolA.UnbondedTokens.Add(sdk.NewInt(10)))

	// same number of unbonded shares / tokens when exchange rate is one
	assert.True(t, poolB.UnbondedShares.Equal(sdk.NewRatFromInt(poolB.UnbondedTokens)))
}

func TestRemoveSharesUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	poolB, tokensB := poolA.removeSharesUnbonded(sdk.NewRat(10))
	assert.Equal(t, poolB.unbondedShareExRate(), sdk.OneRat())

	// correct changes to unbonded shares and bonded pool
	assert.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Sub(sdk.NewRat(10)))
	assert.Equal(t, poolB.UnbondedTokens, poolA.UnbondedTokens.Sub(tokensB))

	// same number of unbonded shares / tokens when exchange rate is one
	assert.True(t, poolB.UnbondedShares.Equal(sdk.NewRatFromInt(poolB.UnbondedTokens)))
}
