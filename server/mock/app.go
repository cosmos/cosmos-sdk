package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/db/badgerdb"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
)

func testTxHandler(options middleware.TxHandlerOptions) tx.Handler {
	return middleware.ComposeMiddlewares(
		middleware.NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		middleware.NewTxDecoderMiddleware(options.TxDecoder),
		middleware.GasTxMiddleware,
		middleware.RecoveryTxMiddleware,
		middleware.NewIndexEventsTxMiddleware(options.IndexEvents),
	)
}

// NewApp creates a simple mock kvstore app for testing. It should work
// similar to a real app. Make sure rootDir is empty before running the test,
// in order to guarantee consistent results
func NewApp(rootDir string, logger log.Logger) (abci.Application, error) {
	db, err := badgerdb.NewDB(filepath.Join(rootDir, "data", "mock"))
	if err != nil {
		return nil, err
	}

	// Capabilities key to access the main KVStore.
	capKeyMainStore := sdk.NewKVStoreKey("main")

	// Create BaseApp.
	opt := bam.SetSubstores(capKeyMainStore)
	baseApp := bam.NewBaseApp("kvstore", logger, db, opt)

	baseApp.SetInitChainer(InitChainer(capKeyMainStore))

	// Set a Route.
	encCfg := simapp.MakeTestEncodingConfig()
	legacyRouter := middleware.NewLegacyRouter()
	// We're adding a test legacy route here, which accesses the kvstore
	// and simply sets the Msg's key/value pair in the kvstore.
	legacyRouter.AddRoute(sdk.NewRoute("kvstore", KVStoreHandler(capKeyMainStore)))
	txHandler := testTxHandler(
		middleware.TxHandlerOptions{
			LegacyRouter:     legacyRouter,
			MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
			TxDecoder:        decodeTx,
		},
	)
	baseApp.SetTxHandler(txHandler)
	if err = baseApp.Init(); err != nil {
		return nil, err
	}
	return baseApp, nil
}

// KVStoreHandler is a simple handler that takes kvstoreTx and writes
// them to the db
func KVStoreHandler(storeKey storetypes.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		dTx, ok := msg.(*kvstoreTx)
		if !ok {
			return nil, errors.New("KVStoreHandler should only receive kvstoreTx")
		}

		// tx is already unmarshalled
		key := dTx.key
		value := dTx.value

		store := ctx.KVStore(storeKey)
		store.Set(key, value)

		any, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}

		return &sdk.Result{
			Log:          fmt.Sprintf("set %s=%s", key, value),
			MsgResponses: []*codectypes.Any{any},
		}, nil
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
func InitChainer(key storetypes.StoreKey) func(sdk.Context, abci.RequestInitChain) abci.ResponseInitChain {
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
func AppGenState(_ *codec.LegacyAmino, _ types.GenesisDoc, _ []json.RawMessage) (appState json.
	RawMessage, err error) {
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

// AppGenStateEmpty returns an empty transaction state for mocking.
func AppGenStateEmpty(_ *codec.LegacyAmino, _ types.GenesisDoc, _ []json.RawMessage) (
	appState json.RawMessage, err error) {
	appState = json.RawMessage(``)
	return
}
