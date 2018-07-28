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
	p2.BondedTokens = sdk.NewDec(3, 0)
	require.False(t, p1.Equal(p2))
}

func TestAddBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewDec(10, 0)
	pool.BondedTokens = sdk.NewDec(10, 0)

	pool = pool.looseTokensToBonded(sdk.NewDec(10, 0))

	require.True(sdk.DecEq(t, sdk.NewDec(20, 0), pool.BondedTokens))
	require.True(sdk.DecEq(t, sdk.NewDec(0, 0), pool.LooseTokens))
}

func TestRemoveBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewDec(10, 0)
	pool.BondedTokens = sdk.NewDec(10, 0)

	pool = pool.bondedTokensToLoose(sdk.NewDec(5, 0))

	require.True(sdk.DecEq(t, sdk.NewDec(5, 0), pool.BondedTokens))
	require.True(sdk.DecEq(t, sdk.NewDec(15, 0), pool.LooseTokens))
}
