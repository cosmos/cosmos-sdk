package baseapp

import (
	"fmt"
	"runtime/debug"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Key to store the header in the DB itself.
// Use the db directly instead of a store to avoid
// conflicts with handlers writing to the store
// and to avoid affecting the Merkle root.
var dbHeaderKey = []byte("header")

// The ABCI application
type BaseApp struct {
	// initialized on creation
	*sdk.App

	router Router // handle any kind of message

	// must be set
	txDecoder   TxDecoder   // unmarshal []byte into sdk.Tx
	anteHandler AnteHandler // ante handler for fee and auth
}

var _ abci.Application = (*BaseApp)(nil)

// Create and name new BaseApp
// NOTE: The db is used to store the version number for now.
func NewApp(name string, cdc *wire.Codec, logger log.Logger, db dbm.DB) *BaseApp {

	cms := store.NewCommitMultiStore(db)

	sdkapp := sdk.NewApp(name, logger, db, cms)
	app := &BaseApp{
		App:         sdkapp,
		router:      NewRouter(),
		txDecoder:   defaultTxDecoder(cdc),
		anteHandler: nil,
	}

	return app
}

// Implements ABCI
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	// Decode the Tx.
	var result sdk.Result
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
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {

	// Decode the Tx.
	var result sdk.Result
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
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	res = app.App.BeginBlock(req)
	// TODO: Chain multiple beginblockers together (from different modules)
	return
}

// Implements ABCI
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	res = app.App.EndBlock(req)
	// TODO: Chain multiple endblockers together (from different modules)
	return
}

// txBytes may be nil in some cases, eg. in tests.
// Also, in the future we may support "internal" transactions.
func (app *BaseApp) runTx(ctx sdk.Context, isCheckTx bool, txBytes []byte, tx Tx) (result sdk.Result) {
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
		err = err.WithDefaultCodespace(sdk.CodespaceRoot)
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
		return sdk.ErrUnknownRequest("Unrecognized Msg type: " + msgType).Result()
	}

	// Get the correct cache
	var msCache sdk.CacheMultiStore
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
func (app *BaseApp) SetTxDecoder(txDecoder TxDecoder) {
	app.txDecoder = txDecoder
}

// Set the txDecoder function
func (app *BaseApp) SetAnteHandler(ah AnteHandler) {
	app.anteHandler = ah
}

// Return the router
func (app *BaseApp) Router() Router { return app.router }

// default custom logic for transaction decoding
func defaultTxDecoder(cdc *wire.Codec) TxDecoder {
	return func(txBytes []byte) (Tx, sdk.Error) {
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
