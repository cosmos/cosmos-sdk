package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondStatus(t *testing.T) {
	require.False(t, sdk.Unbonded.Equal(sdk.Bonded))
	require.False(t, sdk.Unbonded.Equal(sdk.Unbonding))
	require.False(t, sdk.Bonded.Equal(sdk.Unbonding))
	require.Panicsf(t, func() { sdk.BondStatus(0).String() }, "invalid bond status") // nolint:govet
	require.Equal(t, sdk.BondStatusUnbonded, sdk.Unbonded.String())
	require.Equal(t, sdk.BondStatusBonded, sdk.Bonded.String())
	require.Equal(t, sdk.BondStatusUnbonding, sdk.Unbonding.String())
}

func TestTokensToConsensusPower(t *testing.T) {
	require.Equal(t, int64(0), sdk.TokensToConsensusPower(sdk.NewInt(999_999)))
	require.Equal(t, int64(1), sdk.TokensToConsensusPower(sdk.NewInt(1_000_000)))
}
