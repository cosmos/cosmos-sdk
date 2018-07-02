package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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

	pool.BondedTokens = val.PoolShares.Bonded().RoundInt64()
	pool.BondedShares = val.PoolShares.Bonded()

	poolShares := NewBondedShares(sdk.NewRat(50))
	tokens := poolShares.Tokens(pool)
	require.Equal(t, int64(50), tokens.RoundInt64())

	poolShares = NewUnbondingShares(sdk.NewRat(50))
	tokens = poolShares.Tokens(pool)
	require.Equal(t, int64(50), tokens.RoundInt64())

	poolShares = NewUnbondedShares(sdk.NewRat(50))
	tokens = poolShares.Tokens(pool)
	require.Equal(t, int64(50), tokens.RoundInt64())
}
