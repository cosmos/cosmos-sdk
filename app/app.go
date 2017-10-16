package app

import (
	"bytes"
	"fmt"
	"strings"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	sm "github.com/cosmos/cosmos-sdk/state"
	"github.com/cosmos/cosmos-sdk/version"
)

//nolint
const (
	ModuleNameBase = "base"
	ChainKey       = "chain_id"
)

/////////////////////////// Move to SDK ///////

// BaseApp contains a data store and all info needed
// to perform queries and handshakes.
//
// It should be embeded in another struct for CheckTx,
// DeliverTx and initializing state from the genesis.
type BaseApp struct {
	info  *sm.ChainState
	state *Store

	pending []*abci.Validator
	height  uint64
	logger  log.Logger
}

// NewBaseApp creates a data store to handle queries
func NewBaseApp(store *Store, logger log.Logger) *BaseApp {
	return &BaseApp{
		info:   sm.NewChainState(),
		state:  store,
		logger: logger,
	}
}

// GetChainID returns the currently stored chain
func (app *BaseApp) GetChainID() string {
	return app.info.GetChainID(app.state.Committed())
}

// GetState returns the delivertx state, should be removed
func (app *BaseApp) GetState() sm.SimpleDB {
	return app.state.Append()
}

// Logger returns the application base logger
func (app *BaseApp) Logger() log.Logger {
	return app.logger
}

// Info - ABCI
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
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

// SetOption - ABCI
func (app *BaseApp) SetOption(key string, value string) string {
	return "Not Implemented"
}

// Query - ABCI
func (app *BaseApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if len(reqQuery.Data) == 0 {
		resQuery.Log = "Query cannot be zero length"
		resQuery.Code = abci.CodeType_EncodingError
		return
	}

	return app.state.Query(reqQuery)
}

// Commit - ABCI
func (app *BaseApp) Commit() (res abci.Result) {
	// Commit state
	res = app.state.Commit()
	if res.IsErr() {
		cmn.PanicSanity("Error getting hash: " + res.Error())
	}
	return res
}

// InitChain - ABCI
func (app *BaseApp) InitChain(req abci.RequestInitChain) {
	// for _, plugin := range app.plugins.GetList() {
	// 	plugin.InitChain(app.state, validators)
	// }
}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) {
	app.height++
}

// EndBlock - ABCI
// Returns a list of all validator changes made in this block
func (app *BaseApp) EndBlock(height uint64) (res abci.ResponseEndBlock) {
	// TODO: cleanup in case a validator exists multiple times in the list
	res.Diffs = app.pending
	app.pending = nil
	return
}

func (app *BaseApp) AddValChange(diffs []*abci.Validator) {
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
