package simapp

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := context.Background()

	// We export at last height + 1, because that's the height at which
	// CometBFT will start InitChain.
	latestHeight, err := app.LoadLatestHeight()

	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	height := latestHeight + 1
	// if forZeroHeight {
	// 	height = 0
	// 	app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	// }

	genesis, err := app.ExportGenesis(ctx, height)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        genesis,
		Validators:      nil,
		Height:          int64(height),
		ConsensusParams: cmtproto.ConsensusParams{}, // TODO: CometBFT consensus params
	}, err
}
