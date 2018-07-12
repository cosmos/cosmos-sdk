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
	p2.BondedTokens = sdk.NewRat(3)
	require.False(t, p1.Equal(p2))
}

func TestAddBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	pool.BondedTokens = sdk.NewRat(10)

	pool = pool.addBondedTokens(sdk.NewRat(10))

	require.Equal(t, sdk.NewRat(20), pool.BondedTokens)
	require.Equal(t, sdk.NewRat(0), pool.LooseTokens)
}

func TestRemoveBondedTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	pool.BondedTokens = sdk.NewRat(10)

	pool = pool.removeBondedTokens(sdk.NewRat(5))

	require.Equal(t, sdk.NewRat(5), pool.BondedTokens)
	require.Equal(t, sdk.NewRat(15), pool.LooseTokens)
}
