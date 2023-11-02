package runtime

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const ModuleName = "runtime"

// App implements the common methods for a Cosmos SDK-based application
// specific blockchain.
type AppI interface {
	// The assigned name of the app.
	Name() string

	// Application updates every begin block.
	BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error)

	// Application updates every end block.
	EndBlocker(ctx sdk.Context) (sdk.EndBlock, error)

	// Application update at chain (i.e app) initialization.
	InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error)

	// Loads the app at a given height.
	LoadHeight(height int64) error

	// Exports the state of the application for a genesis file.
	ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (types.ExportedApp, error)
}

// AppSimI implements the common methods for a Cosmos SDK-based application
// specific blockchain that chooses to utilize the sdk simulation framework.
type AppSimI interface {
	AppI
	// Helper for the simulation framework.
	SimulationManager() *module.SimulationManager
}
