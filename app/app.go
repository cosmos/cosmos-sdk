package app

import (
	"bytes"
	"fmt"
	"strings"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
	sm "github.com/cosmos/cosmos-sdk/state"
	"github.com/cosmos/cosmos-sdk/version"
)

//nolint
const (
	ModuleNameBase = "base"
	ChainKey       = "chain_id"
)

// Basecoin - The ABCI application
type Basecoin struct {
	info  *sm.ChainState
	state *Store

	handler sdk.Handler
	tick    Ticker

	pending []*abci.Validator
	height  uint64
	logger  log.Logger
}

// Ticker - tick function
type Ticker func(sm.SimpleDB) ([]*abci.Validator, error)

var _ abci.Application = &Basecoin{}

// NewBasecoin - create a new instance of the basecoin application
func NewBasecoin(handler sdk.Handler, store *Store, logger log.Logger) *Basecoin {
	return &Basecoin{
		handler: handler,
		info:    sm.NewChainState(),
		state:   store,
		logger:  logger,
	}
}

// NewBasecoinTick - create a new instance of the basecoin application with tick functionality
func NewBasecoinTick(handler sdk.Handler, store *Store, logger log.Logger, tick Ticker) *Basecoin {
	return &Basecoin{
		handler: handler,
		info:    sm.NewChainState(),
		state:   store,
		logger:  logger,
		tick:    tick,
	}
}

// GetChainID returns the currently stored chain
func (app *Basecoin) GetChainID() string {
	return app.info.GetChainID(app.state.Committed())
}

// GetState is back... please kill me
func (app *Basecoin) GetState() sm.SimpleDB {
	return app.state.Append()
}

// Info - ABCI
func (app *Basecoin) Info(req abci.RequestInfo) abci.ResponseInfo {
	resp := app.state.Info()
	app.logger.Debug("Info",
		"height", resp.LastBlockHeight,
		"hash", fmt.Sprintf("%X", resp.LastBlockAppHash))
	app.height = resp.LastBlockHeight
	return abci.ResponseInfo{
		Data:             fmt.Sprintf("Basecoin v%v", version.Version),
		LastBlockHeight:  resp.LastBlockHeight,
		LastBlockAppHash: resp.LastBlockAppHash,
	}
}

// InitState - used to setup state (was SetOption)
// to be used by InitChain later
func (app *Basecoin) InitState(key string, value string) string {
	module, key := splitKey(key)
	state := app.state.Append()

	if module == ModuleNameBase {
		if key == ChainKey {
			app.info.SetChainID(state, value)
			return "Success"
		}
		return fmt.Sprintf("Error: unknown base option: %s", key)
	}

	log, err := app.handler.InitState(app.logger, state, module, key, value)
	if err == nil {
		return log
	}
	return "Error: " + err.Error()
}

// SetOption - ABCI
func (app *Basecoin) SetOption(key string, value string) string {
	return "Not Implemented"
}

// DeliverTx - ABCI
func (app *Basecoin) DeliverTx(txBytes []byte) abci.Result {
	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		return errors.Result(err)
	}

	ctx := stack.NewContext(
		app.GetChainID(),
		app.height,
		app.logger.With("call", "delivertx"),
	)
	res, err := app.handler.DeliverTx(ctx, app.state.Append(), tx)

	if err != nil {
		return errors.Result(err)
	}
	app.addValChange(res.Diff)
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
		app.height,
		app.logger.With("call", "checktx"),
	)
	res, err := app.handler.CheckTx(ctx, app.state.Check(), tx)

	if err != nil {
		return errors.Result(err)
	}
	return sdk.ToABCI(res)
}

// Query - ABCI
func (app *Basecoin) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if len(reqQuery.Data) == 0 {
		resQuery.Log = "Query cannot be zero length"
		resQuery.Code = abci.CodeType_EncodingError
		return
	}

	return app.state.Query(reqQuery)
}

// Commit - ABCI
func (app *Basecoin) Commit() (res abci.Result) {
	// Commit state
	res = app.state.Commit()
	if res.IsErr() {
		cmn.PanicSanity("Error getting hash: " + res.Error())
	}
	return res
}

// InitChain - ABCI
func (app *Basecoin) InitChain(req abci.RequestInitChain) {
	// for _, plugin := range app.plugins.GetList() {
	// 	plugin.InitChain(app.state, validators)
	// }
}

// BeginBlock - ABCI
func (app *Basecoin) BeginBlock(req abci.RequestBeginBlock) {
	app.height++

	// for _, plugin := range app.plugins.GetList() {
	// 	plugin.BeginBlock(app.state, hash, header)
	// }

	if app.tick != nil {
		diff, err := app.tick(app.state.Append())
		if err != nil {
			panic(err)
		}
		app.addValChange(diff)
	}
}

// EndBlock - ABCI
// Returns a list of all validator changes made in this block
func (app *Basecoin) EndBlock(height uint64) (res abci.ResponseEndBlock) {
	// TODO: cleanup in case a validator exists multiple times in the list
	res.Diffs = app.pending
	app.pending = nil
	return
}

func (app *Basecoin) addValChange(diffs []*abci.Validator) {
	for _, d := range diffs {
		idx := pubKeyIndex(d, app.pending)
		if idx >= 0 {
			app.pending[idx] = d
		} else {
			app.pending = append(app.pending, d)
		}
	}
}

// return index of list with validator of same PubKey, or -1 if no match
func pubKeyIndex(val *abci.Validator, list []*abci.Validator) int {
	for i, v := range list {
		if bytes.Equal(val.PubKey, v.PubKey) {
			return i
		}
	}
	return -1
}

//TODO move split key to tmlibs?

// Splits the string at the first '/'.
// if there are none, assign default module ("base").
func splitKey(key string) (string, string) {
	if strings.Contains(key, "/") {
		keyParts := strings.SplitN(key, "/", 2)
		return keyParts[0], keyParts[1]
	}
	return ModuleNameBase, key
}
