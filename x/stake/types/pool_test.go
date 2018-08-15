package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPoolEqual(t *testing.T) {
	p1 := InitialPool()
	p2 := InitialPool()
	require.True(t, p1.Equal(p2))
	p2.BondedTokens = sdk.NewDec(3)
	require.False(t, p1.Equal(p2))
}

func TestAddBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewDec(10)
	pool.BondedTokens = sdk.NewDec(10)

	pool = pool.looseTokensToBonded(sdk.NewDec(10))

	require.True(sdk.DecEq(t, sdk.NewDec(20), pool.BondedTokens))
	require.True(sdk.DecEq(t, sdk.NewDec(0), pool.LooseTokens))
}

func TestRemoveBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewDec(10)
	pool.BondedTokens = sdk.NewDec(10)

	pool = pool.bondedTokensToLoose(sdk.NewDec(5))

	require.True(sdk.DecEq(t, sdk.NewDec(5), pool.BondedTokens))
	require.True(sdk.DecEq(t, sdk.NewDec(15), pool.LooseTokens))
}
