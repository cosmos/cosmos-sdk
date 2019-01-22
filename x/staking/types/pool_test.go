package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPoolEqual(t *testing.T) {
	p1 := InitialPool()
	p2 := InitialPool()
	require.True(t, p1.Equal(p2))
	p2.BondedTokens = sdk.NewInt(3)
	require.False(t, p1.Equal(p2))
}

func TestAddBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewInt(10)
	pool.BondedTokens = sdk.NewInt(10)

	pool = pool.notBondedTokensToBonded(sdk.NewInt(10))

	require.True(sdk.IntEq(t, sdk.NewInt(20), pool.BondedTokens))
	require.True(sdk.IntEq(t, sdk.NewInt(0), pool.NotBondedTokens))
}

func TestRemoveBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewInt(10)
	pool.BondedTokens = sdk.NewInt(10)

	pool = pool.bondedTokensToNotBonded(sdk.NewInt(5))

	require.True(sdk.IntEq(t, sdk.NewInt(5), pool.BondedTokens))
	require.True(sdk.IntEq(t, sdk.NewInt(15), pool.NotBondedTokens))
}
