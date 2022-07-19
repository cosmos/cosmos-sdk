package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/testslashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

func TestExportAndInitGenesis(t *testing.T) {
	var slashingKeeper slashingkeeper.Keeper
	var stakingKeeper *stakingkeeper.Keeper
	var bankKeeper bankkeeper.Keeper

	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&slashingKeeper,
		&stakingKeeper,
		&bankKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	slashingKeeper.SetParams(ctx, testslashing.TestParams())

	addrDels := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, stakingKeeper.TokensFromConsensusPower(ctx, 200))

	info1 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3),
		time.Now().UTC().Add(100000000000), false, int64(10))
	info2 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4),
		time.Now().UTC().Add(10000000000), false, int64(10))

	slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), info1)
	slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]), info2)
	genesisState := slashingKeeper.ExportGenesis(ctx)

	require.Equal(t, genesisState.Params, testslashing.TestParams())
	require.Len(t, genesisState.SigningInfos, 2)
	require.Equal(t, genesisState.SigningInfos[0].ValidatorSigningInfo, info1)

	// Tombstone validators after genesis shouldn't effect genesis state
	slashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[0]))
	slashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[1]))

	ok := slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0]))
	require.True(t, ok)

	newInfo1, ok := slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	require.NotEqual(t, info1, newInfo1)
	// Initialise genesis with genesis state before tombstone

	slashingKeeper.InitGenesis(ctx, stakingKeeper, genesisState)

	// Validator isTombstoned should return false as GenesisState is initialised
	ok = slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0]))
	require.False(t, ok)

	newInfo1, ok = slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	newInfo2, ok := slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]))
	require.True(t, ok)
	require.Equal(t, info1, newInfo1)
	require.Equal(t, info2, newInfo2)
}
