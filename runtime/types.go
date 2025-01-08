package runtime

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const ModuleName = "runtime"

// AppI implements the common methods for a Cosmos SDK-based application
// specific blockchain.
type AppI interface {
	// Name the assigned name of the app.
	Name() string

	// BeginBlocker updates every begin block.
	BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error)

	// EndBlocker updates every end block.
	EndBlocker(ctx sdk.Context) (sdk.EndBlock, error)

	// InitChainer update at chain (i.e app) initialization.
	InitChainer(ctx sdk.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error)

	// LoadHeight load the app at a given height.
	LoadHeight(height int64) error

	// ExportAppStateAndValidators exports the state of the application for a genesis file.
	ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (ExportedApp, error)
}

// ExportedApp represents an exported app state, along with
// validators, consensus params and latest app height.
type ExportedApp struct {
	// AppState is the application state as JSON.
	AppState json.RawMessage
	// Validators is the exported validator set.
	Validators []sdk.GenesisValidator
	// Height is the app's latest block height.
	Height int64
	// ConsensusParams are the exported consensus params for ABCI.
	ConsensusParams cmtproto.ConsensusParams
}

// AppSimI implements the common methods for a Cosmos SDK-based application
// specific blockchain that chooses to utilize the sdk simulation framework.
type AppSimI interface {
	AppI
	// SimulationManager helper for the simulation framework.
	SimulationManager() *module.SimulationManager
}
