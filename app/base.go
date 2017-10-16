package app

import (
	"fmt"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	handler sdk.Handler
	clock   sdk.Ticker
}

var _ abci.Application = &BaseApp{}

// NewBaseApp extends a StoreApp with a handler and a ticker,
// which it binds to the proper abci calls
func NewBaseApp(store *StoreApp, handler sdk.Handler, clock sdk.Ticker) *BaseApp {
	return &BaseApp{
		StoreApp: store,
		handler:  handler,
		clock:    clock,
	}
}

// DeliverTx - ABCI - dispatches to the handler
func (app *BaseApp) DeliverTx(txBytes []byte) abci.Result {
	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		return errors.Result(err)
	}

	ctx := stack.NewContext(
		app.GetChainID(),
		app.WorkingHeight(),
		app.Logger().With("call", "delivertx"),
	)
	res, err := app.handler.DeliverTx(ctx, app.Append(), tx)

	if err != nil {
		return errors.Result(err)
	}
	app.AddValChange(res.Diff)
	return sdk.ToABCI(res)
}

// CheckTx - ABCI - dispatches to the handler
func (app *BaseApp) CheckTx(txBytes []byte) abci.Result {
	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		return errors.Result(err)
	}

	ctx := stack.NewContext(
		app.GetChainID(),
		app.WorkingHeight(),
		app.Logger().With("call", "checktx"),
	)
	res, err := app.handler.CheckTx(ctx, app.Check(), tx)

	if err != nil {
		return errors.Result(err)
	}
	return sdk.ToABCI(res)
}

// BeginBlock - ABCI - triggers Tick actions
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) {
	// execute tick if present
	if app.clock != nil {
		ctx := stack.NewContext(
			app.GetChainID(),
			app.WorkingHeight(),
			app.Logger().With("call", "tick"),
		)

		diff, err := app.clock.Tick(ctx, app.Append())
		if err != nil {
			panic(err)
		}
		app.AddValChange(diff)
	}
}

// InitState - used to setup state (was SetOption)
// to be used by InitChain later
//
// TODO: rethink this a bit more....
func (app *BaseApp) InitState(module, key, value string) (string, error) {
	state := app.Append()

	if module == ModuleNameBase {
		if key == ChainKey {
			app.info.SetChainID(state, value)
			return "Success", nil
		}
		return "", fmt.Errorf("unknown base option: %s", key)
	}

	log, err := app.handler.InitState(app.Logger(), state, module, key, value)
	if err != nil {
		app.Logger().Error("Genesis App Options", "err", err)
	} else {
		app.Logger().Info(log)
	}
	return log, err
}
