package simapp

import (
	"context"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/x/staking"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(
	jailAllowedAddrs []string,
) (v2.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := context.Background()

	latestHeight, err := app.LoadLatestHeight()
	if err != nil {
		return v2.ExportedApp{}, err
	}

	genesis, err := app.ExportGenesis(ctx, latestHeight)
	if err != nil {
		return v2.ExportedApp{}, err
	}

	_, dbState, err := app.GetStore().StateLatest()
	if err != nil {
		return v2.ExportedApp{}, err
	}
	var validators []sdk.GenesisValidator
	genesisCtx := services.NewGenesisContext(dbState)
	err = genesisCtx.Read(ctx, func(ctx context.Context) error {
		validators, err = staking.WriteValidators(ctx, app.StakingKeeper)
		return err
	})

	return v2.ExportedApp{
		AppState:   genesis,
		Height:     int64(latestHeight),
		Validators: validators,
	}, err
}
