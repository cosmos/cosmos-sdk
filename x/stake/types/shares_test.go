package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPoolSharesTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = 10

	val := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		PoolShares:      NewBondedShares(sdk.NewRat(100)),
		DelegatorShares: sdk.NewRat(100),
	}

	pool.BondedTokens = val.PoolShares.Bonded().Evaluate()
	pool.BondedShares = val.PoolShares.Bonded()

	poolShares := NewBondedShares(sdk.NewRat(50))
	tokens := poolShares.Tokens(pool)
	require.Equal(t, int64(50), tokens.Evaluate())

	poolShares = NewUnbondingShares(sdk.NewRat(50))
	tokens = poolShares.Tokens(pool)
	require.Equal(t, int64(50), tokens.Evaluate())

	poolShares = NewUnbondedShares(sdk.NewRat(50))
	tokens = poolShares.Tokens(pool)
	require.Equal(t, int64(50), tokens.Evaluate())
}
