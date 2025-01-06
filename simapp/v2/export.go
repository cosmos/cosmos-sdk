package simapp

import (
	"context"

	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/x/staking"

	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
// This is a demonstration of how to export a genesis file. Export may need extended at
// the user discretion for cleaning the genesis state at the end provided with jailAllowedAddrs
// Same applies for forZeroHeight preprocessing.
func (app *SimApp[T]) ExportAppStateAndValidators(
	forZeroHeight bool,
	jailAllowedAddrs []string,
) (v2.ExportedApp, error) {
	ctx := context.Background()
	var exportedApp v2.ExportedApp

	latestHeight, err := app.LoadLatestHeight()
	if err != nil {
		return exportedApp, err
	}

	genesis, err := app.ExportGenesis(ctx, latestHeight)
	if err != nil {
		return exportedApp, err
	}

	readerMap, err := app.Store().StateAt(latestHeight)
	if err != nil {
		return exportedApp, err
	}
	genesisCtx := services.NewGenesisContext(readerMap)
	err = genesisCtx.Read(ctx, func(ctx context.Context) error {
		exportedApp.Validators, err = staking.WriteValidators(ctx, app.StakingKeeper)
		return err
	})
	if err != nil {
		return exportedApp, err
	}

	exportedApp.AppState = genesis
	exportedApp.Height = int64(latestHeight)
	if forZeroHeight {
		exportedApp.Height = 0
	}

	return exportedApp, nil
}
