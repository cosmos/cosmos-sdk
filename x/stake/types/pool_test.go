package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPoolEqual(t *testing.T) {
	p1 := InitialPool()
	p2 := InitialPool()

	ok := p1.Equal(p2)
	require.True(t, ok)

	p2.BondedTokens = 3
	p2.BondedShares = sdk.NewRat(10)

	ok = p1.Equal(p2)
	require.False(t, ok)
}

func TestAddBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	pool.BondedTokens = sdk.NewRat(10)

	pool = pool.addTokensBonded(sdk.NewRat(10))

	require.Equal(t, sdk.NewRat(20), pool.BondedTokens)
	require.Equal(t, sdk.NewRat(0), pool.LooseTokens)
}

func TestRemoveBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	pool.BondedTokens = sdk.NewRat(10)

	pool = pool.removeTokensBonded(sdk.NewRat(5))

	require.Equal(t, sdk.NewRat(5), pool.BondedTokens)
	require.Equal(t, sdk.NewRat(15), pool.LooseTokens)
}
