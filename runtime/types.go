package runtime

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const ModuleName = "runtime"

// AppI implements the common methods for a Cosmos SDK-based application
// specific blockchain.
type AppI interface {
	// Name is the assigned name of the app.
	Name() string

	// LegacyAmino is the application types codec.
	// NOTE: This should NOT be sealed before being returned.
	LegacyAmino() *codec.LegacyAmino

	// BeginBlocker is logic run every begin block.
	BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error)

	// EndBlocker is logic run every end block.
	EndBlocker(ctx sdk.Context) (sdk.EndBlock, error)

	// InitChainer is the application update at chain (i.e app) initialization.
	InitChainer(ctx sdk.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error)

	// LoadHeight loads the app at a given height.
	LoadHeight(height int64) error

	// ExportAppStateAndValidators exports the state of the application for a genesis file.
	ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (types.ExportedApp, error)

	// SimulationManager is a helper for the simulation framework.
	SimulationManager() *module.SimulationManager
}
