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
	p2.BondedTokens = sdk.NewUint(3)
	require.False(t, p1.Equal(p2))
}

func TestAddBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewUint(10)
	pool.BondedTokens = sdk.NewUint(10)

	pool = pool.notBondedTokensToBonded(sdk.NewUint(10))

	require.True(sdk.IntEq(t, sdk.NewUint(20), pool.BondedTokens))
	require.True(sdk.IntEq(t, sdk.NewUint(0), pool.NotBondedTokens))
}

func TestRemoveBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewUint(10)
	pool.BondedTokens = sdk.NewUint(10)

	pool = pool.bondedTokensToNotBonded(sdk.NewUint(5))

	require.True(sdk.IntEq(t, sdk.NewUint(5), pool.BondedTokens))
	require.True(sdk.IntEq(t, sdk.NewUint(15), pool.NotBondedTokens))
}
