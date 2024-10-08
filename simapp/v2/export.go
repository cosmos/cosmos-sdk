package simapp

import (
	"context"
	"cosmossdk.io/x/staking"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(
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

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	if err != nil {
		return v2.ExportedApp{}, err
	}

	return v2.ExportedApp{
		AppState:   genesis,
		Height:     int64(latestHeight),
		Validators: validators,
	}, err
}
