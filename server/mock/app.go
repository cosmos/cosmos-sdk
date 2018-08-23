package mock

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	gc "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// NewApp creates a simple mock kvstore app for testing. It should work
// similar to a real app. Make sure rootDir is empty before running the test,
// in order to guarantee consistent results
func NewApp(rootDir string, logger log.Logger) (abci.Application, error) {
	db, err := dbm.NewGoLevelDB("mock", filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, err
	}

	// Capabilities key to access the main KVStore.
	capKeyMainStore := sdk.NewKVStoreKey("main")

	// Create BaseApp.
	baseApp := bam.NewBaseApp("kvstore", nil, logger, db)

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoresIAVL(capKeyMainStore)

	// Set Tx decoder
	baseApp.SetTxDecoder(decodeTx)

	baseApp.SetInitChainer(InitChainer(capKeyMainStore))

	// Set a handler Route.
	baseApp.Router().AddRoute("kvstore", KVStoreHandler(capKeyMainStore))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(capKeyMainStore); err != nil {
		return nil, err
	}

	return baseApp, nil
}

// KVStoreHandler is a simple handler that takes kvstoreTx and writes
// them to the db
func KVStoreHandler(storeKey sdk.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		dTx, ok := msg.(kvstoreTx)
		if !ok {
			panic("KVStoreHandler should only receive kvstoreTx")
		}

		// tx is already unmarshalled
		key := dTx.key
		value := dTx.value

		store := ctx.KVStore(storeKey)
		store.Set(key, value)

		return sdk.Result{
			Code: 0,
			Log:  fmt.Sprintf("set %s=%s", key, value),
		}
	}
}

// basic KV structure
type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// What Genesis JSON is formatted as
type GenesisJSON struct {
	Values []KV `json:"values"`
}

// InitChainer returns a function that can initialize the chain
// with key/value pairs
func InitChainer(key sdk.StoreKey) func(sdk.Context, abci.RequestInitChain) abci.ResponseInitChain {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		stateJSON := req.AppStateBytes

		genesisState := new(GenesisJSON)
		err := json.Unmarshal(stateJSON, genesisState)
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			// return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, val := range genesisState.Values {
			store := ctx.KVStore(key)
			store.Set([]byte(val.Key), []byte(val.Value))
		}
		return abci.ResponseInitChain{}
	}
}

// AppGenState can be passed into InitCmd, returns a static string of a few
// key-values that can be parsed by InitChainer
func AppGenState(_ *wire.Codec, _ []json.RawMessage) (appState json.RawMessage, err error) {
	appState = json.RawMessage(`{
  "values": [
    {
        "key": "hello",
        "value": "goodbye"
    },
    {
        "key": "foo",
        "value": "bar"
    }
  ]
}`)
	return
}

// Return a validator, not much else
func AppGenTx(_ *wire.Codec, pk crypto.PubKey, genTxConfig gc.GenTx) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}
	return
}
