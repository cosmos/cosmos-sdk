package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func testEqualParams(t *testing.T, ctx sdk.Context, params DefaultParams, keeper Keeper) {
	require.Equal(t, params.MaxEvidenceAge, keeper.MaxEvidenceAge(ctx))
	require.Equal(t, params.SignedBlocksWindow, keeper.SignedBlocksWindow(ctx))
	require.Equal(t, sdk.NewRat(params.SignedBlocksWindow).Mul(params.MinSignedPerWindow).RoundInt64(), keeper.MinSignedPerWindow(ctx))
	require.Equal(t, params.DoubleSignUnbondDuration, keeper.DoubleSignUnbondDuration(ctx))
	require.Equal(t, params.DowntimeUnbondDuration, keeper.DowntimeUnbondDuration(ctx))

	require.Equal(t, params.SlashFractionDoubleSign, keeper.SlashFractionDoubleSign(ctx))
	require.Equal(t, params.SlashFractionDowntime, keeper.SlashFractionDowntime(ctx))

}

func TestGenesis(t *testing.T) {
	def := HubDefaultParams()

	ctx, _, _, store, k := createTestInput(t, def)

	state := GenesisState{def}
	err := InitGenesis(ctx, k, state)
	require.Nil(t, err)
	testEqualParams(t, ctx, def, k)

	def.MaxEvidenceAge = 1
	store.Set(ctx, maxEvidenceAgeKey, def.MaxEvidenceAge)
	testEqualParams(t, ctx, def, k)

	def.SignedBlocksWindow = 1
	store.Set(ctx, signedBlocksWindowKey, def.SignedBlocksWindow)
	testEqualParams(t, ctx, def, k)

	def.MinSignedPerWindow = sdk.OneRat()
	store.Set(ctx, minSignedPerWindowKey, def.MinSignedPerWindow)
	testEqualParams(t, ctx, def, k)

	def.DoubleSignUnbondDuration = 1
	store.Set(ctx, doubleSignUnbondDurationKey, def.DoubleSignUnbondDuration)
	testEqualParams(t, ctx, def, k)

	def.DowntimeUnbondDuration = 1
	store.Set(ctx, downtimeUnbondDurationKey, def.DowntimeUnbondDuration)
	testEqualParams(t, ctx, def, k)

	def.SlashFractionDoubleSign = sdk.OneRat()
	store.Set(ctx, slashFractionDoubleSignKey, def.SlashFractionDoubleSign)
	testEqualParams(t, ctx, def, k)

	def.SlashFractionDowntime = sdk.OneRat()
	store.Set(ctx, slashFractionDowntimeKey, def.SlashFractionDowntime)
	testEqualParams(t, ctx, def, k)
}
