package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
)

func TestBonded(t *testing.T) {
	tokenSupply := types.TokensFromConsensusPower(10)
	require.Equal(t, types.NewInt(10000000), tokenSupply)

	bondedTokens := types.TokensFromConsensusPower(1)
	bondedTokens.ToDec()

	require.Equal(t, types.NewInt(10000000), bondedTokens.ToDec().QuoInt(tokenSupply))
}
