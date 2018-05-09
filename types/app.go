package types

import (
	"fmt"
	"runtime/debug"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
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
	*baseapp.BaseApp

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
		BaseApp:     baseapp,
		router:      NewRouter(),
		txDecoder:   defaultTxDecoder(cdc),
		anteHandler: nil,
	}

	return app
}

// Implements ABCI
func (app *App) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	// Decode the Tx.
	var result baseapp.Result
	var tx, err = app.txDecoder(txBytes)
	if err != nil {
		result = err.Result()
	} else {
		result = app.runTx(app.GetCheckContext(), true, txBytes, tx)
	}

	return abci.ResponseCheckTx{
		Code:      uint32(result.Code),
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: result.GasWanted,
		Fee: cmn.KI64Pair{
			[]byte(result.FeeDenom),
			result.FeeAmount,
		},
		Tags: result.Tags,
	}
}

// Implements ABCI
func (app *App) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {

	// Decode the Tx.
	var result baseapp.Result
	var tx, err = app.txDecoder(txBytes)
	if err != nil {
		result = err.Result()
	} else {
		result = app.runTx(app.GetCheckContext(), false, txBytes, tx)
	}

	// After-handler hooks.
	if result.IsOK() {
		app.AppendValUpdates(result.ValidatorUpdates)
	} else {
		// Even though the Code is not OK, there will be some side
		// effects, like those caused by fee deductions or sequence
		// incrementations.
	}

	// Tell the blockchain engine (i.e. Tendermint).
	return abci.ResponseDeliverTx{
		Code:      uint32(result.Code),
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: result.GasWanted,
		GasUsed:   result.GasUsed,
		Tags:      result.Tags,
	}
}

// Implements ABCI
func (app *App) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	res = app.BaseApp.BeginBlock(req)
	// TODO
	return
}

// Implements ABCI
func (app *App) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	res = app.BaseApp.EndBlock(req)
	// TODO
	return
}

// txBytes may be nil in some cases, eg. in tests.
// Also, in the future we may support "internal" transactions.
func (app *App) runTx(ctx baseapp.Context, isCheckTx bool, txBytes []byte, tx Tx) (result baseapp.Result) {
	// Handle any panics.
	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("Recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = ErrInternal(log).Result()
		}
	}()

	// Get the Msg.
	var msg = tx.GetMsg()
	if msg == nil {
		return ErrInternal("Tx.GetMsg() returned nil").Result()
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
		msCache = app.CacheCheckMultiStore()
		ctx = ctx.WithMultiStore(msCache)
	} else {
		// CacheWrap app.deliverState.ms in case it fails.
		msCache = app.CacheDeliverMultiStore()
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
	return func(txBytes []byte) (Tx, baseapp.Error) {
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
		return tx, nil
	}
}
