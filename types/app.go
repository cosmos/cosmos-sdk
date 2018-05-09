package types

import (
	"fmt"
	"runtime/debug"

	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// Key to store the header in the DB itself.
// Use the db directly instead of a store to avoid
// conflicts with handlers writing to the store
// and to avoid affecting the Merkle root.
var dbHeaderKey = []byte("header")

// The ABCI application
type App struct {
	// initialized on creation
	baseapp *baseapp.BaseApp

	router Router // handle any kind of message

	// must be set
	txDecoder   TxDecoder   // unmarshal []byte into sdk.Tx
	anteHandler AnteHandler // ante handler for fee and auth
}

var _ abci.Application = (*App)(nil)

// Create and name new BaseApp
// NOTE: The db is used to store the version number for now.
func NewApp(name string, cdc *wire.Codec, logger log.Logger, db dbm.DB) *App {

	cms := store.NewCommitMultiStore(db)

	baseapp := baseapp.NewBaseApp(name, logger, db, cms)

	app := &App{
		baseapp:   baseapp,
		router:    NewRouter(),
		txDecoder: defaultTxDecoder(cdc),
	}

	store.NewCommitMultiStore(db)

	baseapp.SetCheckTxer(checkTxer)
	baseapp.SetDeliverTxer(deliverTxer)

	return app
}

var _ baseapp.CheckTxer = checkTxer

func checkTxer(ctx baseapp.Context, txBytes []byte) (result baseapp.Result) {
	var tx, err = app.txDecoder(txBytes)
	if err != nil {
		result = err.Result()
	} else {
		result = runTx(ctx, true, txBytes, tx)
	}
	return
}

var _ baseapp.DeliverTxer = deliverTxer

func deliverTxer(ctx baseapp.Context, txBytes []byte) (result baseapp.Result) {
	// Decode the Tx.
	var tx, err = app.txDecoder(txBytes)
	if err != nil {
		result = err.Result()
	} else {
		result = runTx(ctx, false, txBytes, tx)
	}
	return
}

// txBytes may be nil in some cases, eg. in tests.
// Also, in the future we may support "internal" transactions.
func runTx(ctx baseapp.Context, isCheckTx bool, txBytes []byte, tx Tx) (result baseapp.Result) {
	// Handle any panics.
	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("Recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = sdk.ErrInternal(log).Result()
		}
	}()

	// Get the Msg.
	var msg = tx.GetMsg()
	if msg == nil {
		return baseapp.ErrInternal("Tx.GetMsg() returned nil").Result()
	}

	// Validate the Msg.
	err := msg.ValidateBasic()
	if err != nil {
		err = err.WithDefaultCodespace(baseapp.CodespaceRoot)
		return err.Result()
	}

	// Run the ante handler.
	if app.anteHandler != nil {
		newCtx, result, abort := app.anteHandler(ctx, tx)
		if abort {
			return result
		}
		if !newCtx.IsZero() {
			ctx = newCtx
		}
	}

	// Match route.
	msgType := msg.Type()
	handler := app.router.Route(msgType)
	if handler == nil {
		return baseapp.ErrUnknownRequest("Unrecognized Msg type: " + msgType).Result()
	}

	// Get the correct cache
	var msCache baseapp.CacheMultiStore
	if isCheckTx == true {
		// CacheWrap app.checkState.ms in case it fails.
		msCache = app.checkState.CacheMultiStore()
		ctx = ctx.WithMultiStore(msCache)
	} else {
		// CacheWrap app.deliverState.ms in case it fails.
		msCache = app.deliverState.CacheMultiStore()
		ctx = ctx.WithMultiStore(msCache)

	}

	result = handler(ctx, msg)

	// If result was successful, write to app.checkState.ms or app.deliverState.ms
	if result.IsOK() {
		msCache.Write()
	}

	return result
}

// Set the txDecoder function
func (app *App) SetTxDecoder(txDecoder TxDecoder) {
	app.txDecoder = txDecoder
}

// Set the txDecoder function
func (app *App) SetAnteHandler(ah AnteHandler) {
	app.anteHandler = ah
}

// Return the router
func (app *App) Router() Router { return app.router }

// default custom logic for transaction decoding
func defaultTxDecoder(cdc *wire.Codec) TxDecoder {
	return func(txBytes []byte) (Tx, Error) {
		var tx Tx

		if len(txBytes) == 0 {
			return nil, ErrTxDecode("txBytes are empty")
		}

		// StdTx.Msg is an interface. The concrete types
		// are registered by MakeTxCodec
		err := cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, ErrTxDecode("") // TODO: FIX WITH ERRORS:   .Trace(err.Error())
		}
		return tx, Error{} // TODO: Was  nil
	}
}
