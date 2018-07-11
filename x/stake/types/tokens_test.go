package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	tokens := NewBondedTokens(sdk.NewRat(50))
	require.Equal(t, int64(50), tokens.Bonded().RoundInt64())

	tokens = NewUnbondingTokens(sdk.NewRat(50))
	require.Equal(t, int64(50), tokens.Unbonding().RoundInt64())

	tokens = NewUnbondedTokens(sdk.NewRat(50))
	require.Equal(t, int64(50), tokens.Unbonded().RoundInt64())
}
