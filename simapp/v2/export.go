package simapp

import (
	"context"
	"cosmossdk.io/x/staking"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := context.Background()

	latestHeight, err := app.LoadLatestHeight()

	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	height := latestHeight

	genesis, err := app.ExportGenesis(ctx, latestHeight)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	var validators []cmttypes.GenesisValidator
	_, err = app.RunWithCtx(ctx, func(ctx context.Context) error {
		validators, err = staking.WriteValidators(ctx, app.StakingKeeper)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        genesis,
		Validators:      validators,
		Height:          int64(height),
		ConsensusParams: cmtproto.ConsensusParams{}, // TODO: CometBFT consensus params
	}, err
}
