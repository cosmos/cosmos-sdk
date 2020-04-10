package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInitGenesis(t *testing.T) {
	cdc, ctx, _, stakingKeeper, _, slashingKeeper := createTestInput(t, DefaultParams())
	appModule := NewAppModule(slashingKeeper, stakingKeeper)
	var exportState GenesisState

	// 1.check default export
	cdc.MustUnmarshalJSON(appModule.ExportGenesis(ctx), &exportState)
	require.Equal(t, DefaultGenesisState(), exportState)

	// 2.change params and check again
	initParams := NewParams(10000000000,1000,sdk.MustNewDecFromStr("0.05"), 600000000000,sdk.ZeroDec(),sdk.ZeroDec())
	genesisState := NewGenesisState(initParams, exportState.SigningInfos, exportState.MissedBlocks)
	appModule.InitGenesis(ctx, cdc.MustMarshalJSON(genesisState))

	cdc.MustUnmarshalJSON(appModule.ExportGenesis(ctx), &exportState)
	require.Equal(t, genesisState, exportState)

	// 3.change the state.SigningInfos and state.MissedBlocks info and check again
	conAddress := sdk.GetConsAddress(pks[0])
	slashingKeeper.addPubkey(ctx, pks[0])
	sigingInfo := NewValidatorSigningInfo(conAddress, 10, 1, time.Now().Add(10000), true, 5, types.Destroying)
	slashingKeeper.SetValidatorSigningInfo(ctx, conAddress, sigingInfo)
	slashingKeeper.HandleValidatorSignature(ctx, pks[0].Address(), 100,false)

	cdc.MustUnmarshalJSON(appModule.ExportGenesis(ctx), &exportState)
	require.Equal(t,1, len(exportState.SigningInfos))
	require.Equal(t, 1 ,len(exportState.MissedBlocks))
}