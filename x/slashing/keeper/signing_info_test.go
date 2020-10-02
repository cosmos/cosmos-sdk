package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestGetSetValidatorSigningInfo(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.TokensFromConsensusPower(200))

	info, found := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	require.False(t, found)
	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(addrDels[0]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), newInfo)
	info, found = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	require.True(t, found)
	require.Equal(t, info.StartHeight, int64(4))
	require.Equal(t, info.IndexOffset, int64(3))
	require.Equal(t, info.JailedUntil, time.Unix(2, 0).UTC())
	require.Equal(t, info.MissedBlocksCounter, int64(10))
}

func TestGetSetValidatorMissedBlockBitArray(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.TokensFromConsensusPower(200))

	missed := app.SlashingKeeper.GetValidatorMissedBlockBitArray(ctx, sdk.ConsAddress(addrDels[0]), 0)
	require.False(t, missed) // treat empty key as not missed
	app.SlashingKeeper.SetValidatorMissedBlockBitArray(ctx, sdk.ConsAddress(addrDels[0]), 0, true)
	missed = app.SlashingKeeper.GetValidatorMissedBlockBitArray(ctx, sdk.ConsAddress(addrDels[0]), 0)
	require.True(t, missed) // now should be missed
}

func TestTombstoned(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.TokensFromConsensusPower(200))

	require.Panics(t, func() { app.SlashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[0])) })
	require.False(t, app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0])))

	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(addrDels[0]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), newInfo)

	require.False(t, app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0])))
	app.SlashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[0]))
	require.True(t, app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0])))
	require.Panics(t, func() { app.SlashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[0])) })
}

func TestJailUntil(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.TokensFromConsensusPower(200))

	require.Panics(t, func() { app.SlashingKeeper.JailUntil(ctx, sdk.ConsAddress(addrDels[0]), time.Now()) })

	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(addrDels[0]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), newInfo)
	app.SlashingKeeper.JailUntil(ctx, sdk.ConsAddress(addrDels[0]), time.Unix(253402300799, 0).UTC())

	info, ok := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	require.True(t, ok)
	require.Equal(t, time.Unix(253402300799, 0).UTC(), info.JailedUntil)
}
