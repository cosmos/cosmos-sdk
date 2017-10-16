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
func (app *Basecoin) InitState(module, key, value string) (string, error) {
	state := app.Append()

	if module == ModuleNameBase {
		if key == ChainKey {
			app.info.SetChainID(state, value)
			return "Success", nil
		}
		return "", fmt.Errorf("unknown base option: %s", key)
	}

	return app.handler.InitState(app.Logger(), state, module, key, value)
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

// LoadGenesis parses the genesis file and sets the initial
// state based on that
func (app *Basecoin) LoadGenesis(filePath string) error {
	init, err := GetInitialState(filePath)
	if err != nil {
		return err
	}

	// execute all the genesis init options
	// abort on any error
	fmt.Printf("%#v\n", init)
	for _, mkv := range init {
		log, _ := app.InitState(mkv.Module, mkv.Key, mkv.Value)
		// TODO: error out on bad options??
		// if err != nil {
		// 	return err
		// }
		app.Logger().Info(log)
	}
	return nil
}
