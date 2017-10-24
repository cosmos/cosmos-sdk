package app

import (
	"fmt"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk"
)

// InitApp - The ABCI application with initialization hooks
type InitApp struct {
	*BaseApp
	initState sdk.InitStater
	initVals  sdk.InitValidator
}

var _ abci.Application = &InitApp{}

// NewInitApp extends a BaseApp with initialization callbacks,
// which it binds to the proper abci calls
func NewInitApp(base *BaseApp, initState sdk.InitStater,
	initVals sdk.InitValidator) *InitApp {

	return &InitApp{
		BaseApp:   base,
		initState: initState,
		initVals:  initVals,
	}
}

// InitState - used to setup state (was SetOption)
// to be call from setting up the genesis file
func (app *InitApp) InitState(module, key, value string) error {
	state := app.Append()
	logger := app.Logger().With("module", module, "key", key)

	if module == sdk.ModuleNameBase {
		if key == sdk.ChainKey {
			app.info.SetChainID(state, value)
			return nil
		}
		logger.Error("Invalid genesis option")
		return fmt.Errorf("Unknown base option: %s", key)
	}

	log, err := app.initState.InitState(logger, state, module, key, value)
	if err != nil {
		logger.Error("Invalid genesis option", "err", err)
	} else {
		logger.Info(log)
	}
	return err
}

// InitChain - ABCI - sets the initial validators
func (app *InitApp) InitChain(req abci.RequestInitChain) {
	// return early if no InitValidator registered
	if app.initVals == nil {
		return
	}

	logger, store := app.Logger(), app.Append()
	app.initVals.InitValidators(logger, store, req.Validators)
}
