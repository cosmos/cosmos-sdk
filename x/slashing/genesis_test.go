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

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.TokensFromConsensusPower(200))

	info1 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3),
		time.Unix(2, 0), false, int64(10))
	info2 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4),
		time.Unix(2, 0), true, int64(10))

	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), info1)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]), info2)
	genesisState := slashing.ExportGenesis(ctx, app.SlashingKeeper)

	assert.Equal(t, genesisState.Params, keeper.TestParams())
	assert.Len(t, genesisState.SigningInfos, 2)
	assert.Equal(t, genesisState.SigningInfos[0].ValidatorSigningInfo, info1)
}
