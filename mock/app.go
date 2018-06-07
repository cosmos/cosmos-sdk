package mock

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
	"io/ioutil"
	"os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// Extended ABCI application
type MockApp struct {
	*bam.BaseApp
	Cdc             *wire.Codec // public since the codec is passed into the module anyways.
	KeyMain         *sdk.KVStoreKey
	KeyAccountStore *sdk.KVStoreKey
	// TODO: Abstract this out from not needing to be auth specifically
	AccountMapper       auth.AccountMapper
	FeeCollectionKeeper auth.FeeCollectionKeeper
}

// NewApp is used for testing the server. For the internal mock app stuff, it uses code in helpers.go
func NewApp(rootDir string, logger log.Logger) (abci.Application, error) {
	db, err := dbm.NewGoLevelDB("mock", filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, err
	}

	// Capabilities key to access the main KVStore.
	capKeyMainStore := sdk.NewKVStoreKey("mock")

	// Create MockApp.
	mApp := &MockApp{
		BaseApp:         bam.NewBaseApp("mockApp", nil, logger, db),
		Cdc:             wire.NewCodec(),
		KeyMain:         capKeyMainStore,
		KeyAccountStore: sdk.NewKVStoreKey("acc"),
	}

	// Define the accountMapper.
	mApp.AccountMapper = auth.NewAccountMapper(
		mApp.Cdc,
		mApp.KeyAccountStore, // target store
		&auth.BaseAccount{},  // prototype
	)

	// Set mounts for BaseApp's MultiStore.
	mApp.MountStoresIAVL(capKeyMainStore)

	// Set Tx decoder
	mApp.SetTxDecoder(decodeTx)

	mApp.SetInitChainer(InitChainer(capKeyMainStore))

	// Set a handler Route.
	// TODO:
	mApp.Router().AddRoute("kvstore", KVStoreHandler(capKeyMainStore))

	// Load latest version.
	if err := mApp.LoadLatestVersion(capKeyMainStore); err != nil {
		return nil, err
	}

	return mApp, nil
}

// SetupApp returns an application as well as a clean-up function
// to be used to quickly setup a test case with an app
func SetupApp() (*MockApp, func(), error) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "mock")
	rootDir, err := ioutil.TempDir("", "mock-sdk")
	if err != nil {
		return &MockApp{}, nil, err
	}

	cleanup := func() {
		os.RemoveAll(rootDir)
	}

	app, err := NewApp(rootDir, logger)
	return app.(*MockApp), cleanup, err
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
func AppGenTx(_ *wire.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}
	return
}
