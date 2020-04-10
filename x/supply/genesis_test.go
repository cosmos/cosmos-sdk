package supply

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/keeper"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitGenesis(t *testing.T) {
	_, ctx, accKeeper, supplyKeeper := keeper.CreateTestInput(t, false, 100,2)
	appModule := NewAppModule(supplyKeeper, accKeeper)

	// 1.check default export
	require.Equal(t, `{"supply":[{"denom":"okt","amount":"20000000000.00000000"}]}`,
		string(appModule.ExportGenesis(ctx)))

	// 2.init again && check
	newCdc, newCtx, newAccKeeper, newSupplyKeeper := keeper.CreateTestInput(t, false, 100,2)
	newAppModule := NewAppModule(newSupplyKeeper, newAccKeeper)

	coins := sdk.Coins{{"okt",sdk.NewDec(100000)},{"usd", sdk.NewDec(1000)}}
	genesisState := GenesisState{coins}
	newAppModule.InitGenesis(newCtx, newCdc.MustMarshalJSON(genesisState))

	var exportGenesisState GenesisState
	newCdc.MustUnmarshalJSON(newAppModule.ExportGenesis(newCtx), &exportGenesisState)
	require.Equal(t, genesisState, exportGenesisState)

	// 3. mint coins
	addCoins := sdk.Coins{{"rmb", sdk.NewDec(1000)}}
	genesisState.Supply = genesisState.Supply.Add(addCoins)

	err := newSupplyKeeper.MintCoins(newCtx, "minter", addCoins)
	require.Nil(t, err)

	newCdc.MustUnmarshalJSON(newAppModule.ExportGenesis(newCtx), &exportGenesisState)
	require.Equal(t, genesisState, exportGenesisState)
}
