package app

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
	sm "github.com/cosmos/cosmos-sdk/state"
)

// Basecoin - The ABCI application
type Basecoin struct {
	*BaseApp
	handler sdk.Handler
	tick    Ticker
}

// Ticker - tick function
type Ticker func(sm.SimpleDB) ([]*abci.Validator, error)

var _ abci.Application = &Basecoin{}

// NewBasecoin - create a new instance of the basecoin application
func NewBasecoin(handler sdk.Handler, dbName string, cacheSize int, logger log.Logger) (*Basecoin, error) {
	base, err := NewBaseApp(dbName, cacheSize, logger)
	app := &Basecoin{
		BaseApp: base,
		handler: handler,
	}
	return app, err
}

// NewBasecoinTick - create a new instance of the basecoin application with tick functionality
func NewBasecoinTick(handler sdk.Handler, tick Ticker, dbName string, cacheSize int, logger log.Logger) (*Basecoin, error) {
	base, err := NewBaseApp(dbName, cacheSize, logger)
	app := &Basecoin{
		BaseApp: base,
		handler: handler,
		tick:    tick,
	}
	return app, err
}

// InitState - used to setup state (was SetOption)
// to be used by InitChain later
func (app *Basecoin) InitState(key string, value string) string {
	module, key := splitKey(key)
	state := app.Append()

	if module == ModuleNameBase {
		if key == ChainKey {
			app.info.SetChainID(state, value)
			return "Success"
		}
		return fmt.Sprintf("Error: unknown base option: %s", key)
	}

	log, err := app.handler.InitState(app.Logger(), state, module, key, value)
	if err == nil {
		return log
	}
	return "Error: " + err.Error()
}

// DeliverTx - ABCI
func (app *Basecoin) DeliverTx(txBytes []byte) abci.Result {
	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		return errors.Result(err)
	}

	ctx := stack.NewContext(
		app.GetChainID(),
		app.height+1,
		app.Logger().With("call", "delivertx"),
	)
	res, err := app.handler.DeliverTx(ctx, app.Append(), tx)

	if err != nil {
		return errors.Result(err)
	}
	app.AddValChange(res.Diff)
	return sdk.ToABCI(res)
}

// CheckTx - ABCI
func (app *Basecoin) CheckTx(txBytes []byte) abci.Result {
	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		return errors.Result(err)
	}

	ctx := stack.NewContext(
		app.GetChainID(),
		app.height+1,
		app.Logger().With("call", "checktx"),
	)
	res, err := app.handler.CheckTx(ctx, app.Check(), tx)

	if err != nil {
		return errors.Result(err)
	}
	return sdk.ToABCI(res)
}

// BeginBlock - ABCI
func (app *Basecoin) BeginBlock(req abci.RequestBeginBlock) {
	// call the embeded Begin
	app.BaseApp.BeginBlock(req)

	// now execute tick
	if app.tick != nil {
		diff, err := app.tick(app.Append())
		if err != nil {
			panic(err)
		}
		app.AddValChange(diff)
	}
}
