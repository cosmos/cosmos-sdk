package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondedRatio(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = 1
	pool.BondedTokens = 2

	// bonded pool / total supply
	require.Equal(t, pool.BondedRatio(), sdk.NewRat(2).Quo(sdk.NewRat(3)))

	// avoids divide-by-zero
	pool.LooseTokens = 0
	pool.BondedTokens = 0
	require.Equal(t, pool.BondedRatio(), sdk.ZeroRat())
}

func TestBondedShareExRate(t *testing.T) {
	pool := InitialPool()
	pool.BondedTokens = 3
	pool.BondedShares = sdk.NewRat(10)

	// bonded pool / bonded shares
	require.Equal(t, pool.BondedShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.BondedShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.BondedShareExRate(), sdk.OneRat())
}

func TestUnbondingShareExRate(t *testing.T) {
	pool := InitialPool()
	pool.UnbondingTokens = 3
	pool.UnbondingShares = sdk.NewRat(10)

	// unbonding pool / unbonding shares
	require.Equal(t, pool.UnbondingShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.UnbondingShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.UnbondingShareExRate(), sdk.OneRat())
}

func TestUnbondedShareExRate(t *testing.T) {
	pool := InitialPool()
	pool.UnbondedTokens = 3
	pool.UnbondedShares = sdk.NewRat(10)

	// unbonded pool / unbonded shares
	require.Equal(t, pool.UnbondedShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.UnbondedShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.UnbondedShareExRate(), sdk.OneRat())
}

func TestAddTokensBonded(t *testing.T) {

	poolA := InitialPool()
	poolA.LooseTokens = 10
	require.Equal(t, poolA.BondedShareExRate(), sdk.OneRat())
	poolB, sharesB := poolA.addTokensBonded(10)
	require.Equal(t, poolB.BondedShareExRate(), sdk.OneRat())

	// correct changes to bonded shares and bonded pool
	require.Equal(t, poolB.BondedShares, poolA.BondedShares.Add(sharesB.Amount))
	require.Equal(t, poolB.BondedTokens, poolA.BondedTokens+10)

	// same number of bonded shares / tokens when exchange rate is one
	require.True(t, poolB.BondedShares.Equal(sdk.NewRat(poolB.BondedTokens)))
}

func TestRemoveSharesBonded(t *testing.T) {

	poolA := InitialPool()
	poolA.LooseTokens = 10
	require.Equal(t, poolA.BondedShareExRate(), sdk.OneRat())
	poolB, tokensB := poolA.removeSharesBonded(sdk.NewRat(10))
	require.Equal(t, poolB.BondedShareExRate(), sdk.OneRat())

	// correct changes to bonded shares and bonded pool
	require.Equal(t, poolB.BondedShares, poolA.BondedShares.Sub(sdk.NewRat(10)))
	require.Equal(t, poolB.BondedTokens, poolA.BondedTokens-tokensB)

	// same number of bonded shares / tokens when exchange rate is one
	require.True(t, poolB.BondedShares.Equal(sdk.NewRat(poolB.BondedTokens)))
}

func TestAddTokensUnbonded(t *testing.T) {

	poolA := InitialPool()
	poolA.LooseTokens = 10
	require.Equal(t, poolA.UnbondedShareExRate(), sdk.OneRat())
	poolB, sharesB := poolA.addTokensUnbonded(10)
	require.Equal(t, poolB.UnbondedShareExRate(), sdk.OneRat())

	// correct changes to unbonded shares and unbonded pool
	require.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Add(sharesB.Amount))
	require.Equal(t, poolB.UnbondedTokens, poolA.UnbondedTokens+10)

	// same number of unbonded shares / tokens when exchange rate is one
	require.True(t, poolB.UnbondedShares.Equal(sdk.NewRat(poolB.UnbondedTokens)))
}

func TestRemoveSharesUnbonded(t *testing.T) {

	poolA := InitialPool()
	poolA.UnbondedTokens = 10
	poolA.UnbondedShares = sdk.NewRat(10)
	require.Equal(t, poolA.UnbondedShareExRate(), sdk.OneRat())
	poolB, tokensB := poolA.removeSharesUnbonded(sdk.NewRat(10))
	require.Equal(t, poolB.UnbondedShareExRate(), sdk.OneRat())

	// correct changes to unbonded shares and bonded pool
	require.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Sub(sdk.NewRat(10)))
	require.Equal(t, poolB.UnbondedTokens, poolA.UnbondedTokens-tokensB)

	// same number of unbonded shares / tokens when exchange rate is one
	require.True(t, poolB.UnbondedShares.Equal(sdk.NewRat(poolB.UnbondedTokens)))
}
