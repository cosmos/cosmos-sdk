package slashing_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestExportAndInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.TokensFromConsensusPower(200))

	info1 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3),
		time.Now().UTC().Add(100000000000), false, int64(10))
	info2 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4),
		time.Now().UTC().Add(10000000000), false, int64(10))

	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), info1)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]), info2)
	genesisState := slashing.ExportGenesis(ctx, app.SlashingKeeper)

	assert.Equal(t, genesisState.Params, keeper.TestParams())
	assert.Len(t, genesisState.SigningInfos, 2)
	assert.Equal(t, genesisState.SigningInfos[0].ValidatorSigningInfo, info1)

	// Tombstone validators after genesis shouldn't effect genesis state
	app.SlashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[0]))
	app.SlashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[1]))

	ok := app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0]))
	assert.True(t, ok)

	newInfo1, ok := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	assert.NotEqual(t, info1, newInfo1)
	// Initialise genesis with genesis state before tombstone
	slashing.InitGenesis(ctx, app.SlashingKeeper, app.StakingKeeper, genesisState)

	// Validator isTombstoned should return false as GenesisState is initialised
	ok = app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0]))
	assert.False(t, ok)

	newInfo1, ok = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	newInfo2, ok := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]))
	assert.True(t, ok)
	assert.Equal(t, info1, newInfo1)
	assert.Equal(t, info2, newInfo2)
}
