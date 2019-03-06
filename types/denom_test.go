package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterDenom(t *testing.T) {
	require.Error(t, RegisterDenom(Uatom, ZeroInt()))

	unit := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil))
	require.NoError(t, RegisterDenom("gwei", unit))

	res, ok := GetDenomUnit("gwei")
	require.True(t, ok)
	require.Equal(t, unit, res)

	res, ok = GetDenomUnit("finney")
	require.False(t, ok)
	require.Equal(t, ZeroInt(), res)
}
