package simapp

import (
	"context"

	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(jailAllowedAddrs []string) (v2.ExportedApp, error) {
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

	return v2.ExportedApp{
		AppState: genesis,
		Height:   int64(latestHeight),
	}, nil
}
