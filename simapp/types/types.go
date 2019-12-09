/*
Package types defines the interfaces and types required for an SDK-based app.
*/
package types

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// App implements the common methods for a Cosmos SDK-based application
// specific blockchain.
type App interface {
	GetBaseApp() *baseapp.BaseApp

	Name() string
	Codec() *codec.Codec
	GetKey(storeKey string) *sdk.KVStoreKey
	GetTKey(storeKey string) *sdk.TransientStoreKey
	GetSubspace(moduleName string) params.Subspace

	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock
	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock
	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain
	LoadHeight(height int64) error

	ExportAppStateAndValidators(
		forZeroHeight bool, jailWhiteList []string,
	) (json.RawMessage, []tmtypes.GenesisValidator, error)

	ModuleAccountAddrs() map[string]bool
	SimulationManager() *module.SimulationManager
}
