package app

import (
	"fmt"
	"strings"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	eyes "github.com/tendermint/merkleeyes/client"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
	sm "github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/version"
)

//nolint
const (
	ModuleNameBase = "base"
	ChainKey       = "chain_id"
)

// Basecoin - The ABCI application
type Basecoin struct {
	eyesCli    *eyes.Client
	state      *sm.State
	cacheState *sm.State
	handler    basecoin.Handler
	height     uint64
	logger     log.Logger
}

var _ abci.Application = &Basecoin{}

// NewBasecoin - create a new instance of the basecoin application
func NewBasecoin(handler basecoin.Handler, eyesCli *eyes.Client, logger log.Logger) *Basecoin {
	state := sm.NewState(eyesCli, logger.With("module", "state"))

	return &Basecoin{
		handler:    handler,
		eyesCli:    eyesCli,
		state:      state,
		cacheState: nil,
		height:     0,
		logger:     logger,
	}
}

// DefaultHandler - placeholder to just handle sendtx
func DefaultHandler() basecoin.Handler {
	// use the default stack
	h := coin.NewHandler()
	d := stack.NewDispatcher(stack.WrapHandler(h))
	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
	).Use(d)
}

// GetState - XXX For testing, not thread safe!
func (app *Basecoin) GetState() *sm.State {
	return app.state.CacheWrap()
}

// Info - ABCI
func (app *Basecoin) Info() abci.ResponseInfo {
	resp, err := app.eyesCli.InfoSync()
	if err != nil {
		cmn.PanicCrisis(err)
	}
	app.height = resp.LastBlockHeight
	return abci.ResponseInfo{
		Data:             fmt.Sprintf("Basecoin v%v", version.Version),
		LastBlockHeight:  resp.LastBlockHeight,
		LastBlockAppHash: resp.LastBlockAppHash,
	}
}

// SetOption - ABCI
func (app *Basecoin) SetOption(key string, value string) string {

	module, key := splitKey(key)

	if module == ModuleNameBase {
		if key == ChainKey {
			app.state.SetChainID(value)
			return "Success"
		}
		return fmt.Sprintf("Error: unknown base option: %s", key)
	}

	log, err := app.handler.SetOption(app.logger, app.state, module, key, value)
	if err == nil {
		return log
	}
	return "Error: " + err.Error()
}

// DeliverTx - ABCI
func (app *Basecoin) DeliverTx(txBytes []byte) abci.Result {
	tx, err := basecoin.LoadTx(txBytes)
	if err != nil {
		return errors.Result(err)
	}

	// TODO: can we abstract this setup and commit logic??
	cache := app.state.CacheWrap()
	ctx := stack.NewContext(
		app.state.GetChainID(),
		app.height,
		app.logger.With("call", "delivertx"),
	)
	res, err := app.handler.DeliverTx(ctx, cache, tx)

	if err != nil {
		// discard the cache...
		return errors.Result(err)
	}
	// commit the cache and return result
	cache.CacheSync()
	return res.ToABCI()
}

// CheckTx - ABCI
func (app *Basecoin) CheckTx(txBytes []byte) abci.Result {
	tx, err := basecoin.LoadTx(txBytes)
	if err != nil {
		return errors.Result(err)
	}

	// TODO: can we abstract this setup and commit logic??
	ctx := stack.NewContext(
		app.state.GetChainID(),
		app.height,
		app.logger.With("call", "checktx"),
	)
	// checktx generally shouldn't touch the state, but we don't care
	// here on the framework level, since the cacheState is thrown away next block
	res, err := app.handler.CheckTx(ctx, app.cacheState, tx)

	if err != nil {
		return errors.Result(err)
	}
	return res.ToABCI()
}

// Query - ABCI
func (app *Basecoin) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if len(reqQuery.Data) == 0 {
		resQuery.Log = "Query cannot be zero length"
		resQuery.Code = abci.CodeType_EncodingError
		return
	}

	resQuery, err := app.eyesCli.QuerySync(reqQuery)
	if err != nil {
		resQuery.Log = "Failed to query MerkleEyes: " + err.Error()
		resQuery.Code = abci.CodeType_InternalError
		return
	}
	return
}

// Commit - ABCI
func (app *Basecoin) Commit() (res abci.Result) {

	// Commit state
	res = app.state.Commit()

	// Wrap the committed state in cache for CheckTx
	app.cacheState = app.state.CacheWrap()

	if res.IsErr() {
		cmn.PanicSanity("Error getting hash: " + res.Error())
	}
	return res
}

// InitChain - ABCI
func (app *Basecoin) InitChain(validators []*abci.Validator) {
	// for _, plugin := range app.plugins.GetList() {
	// 	plugin.InitChain(app.state, validators)
	// }
}

// BeginBlock - ABCI
func (app *Basecoin) BeginBlock(hash []byte, header *abci.Header) {
	app.height++
	// for _, plugin := range app.plugins.GetList() {
	// 	plugin.BeginBlock(app.state, hash, header)
	// }
}

// EndBlock - ABCI
func (app *Basecoin) EndBlock(height uint64) (res abci.ResponseEndBlock) {
	// for _, plugin := range app.plugins.GetList() {
	// 	pluginRes := plugin.EndBlock(app.state, height)
	// 	res.Diffs = append(res.Diffs, pluginRes.Diffs...)
	// }
	return
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
