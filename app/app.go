package app

import (
	"fmt"
	"strings"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/modules/roles"
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
	info *sm.ChainState

	state *Store

	handler basecoin.Handler
	height  uint64
	logger  log.Logger
}

var _ abci.Application = &Basecoin{}

// NewBasecoin - create a new instance of the basecoin application
func NewBasecoin(handler basecoin.Handler, store *Store, logger log.Logger) *Basecoin {
	return &Basecoin{
		handler: handler,
		info:    sm.NewChainState(),
		state:   store,
		logger:  logger,
	}
}

// DefaultHandler - placeholder to just handle sendtx
func DefaultHandler(feeDenom string) basecoin.Handler {
	// use the default stack
	c := coin.NewHandler()
	r := roles.NewHandler()
	d := stack.NewDispatcher(
		stack.WrapHandler(c),
		stack.WrapHandler(r),
	)
	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
		stack.Checkpoint{OnCheck: true},
		nonce.ReplayCheck{},
		roles.NewMiddleware(),
		fee.NewSimpleFeeMiddleware(coin.Coin{feeDenom, 0}, fee.Bank),
		stack.Checkpoint{OnDeliver: true},
	).Use(d)
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
func (app *Basecoin) Info() abci.ResponseInfo {
	resp := app.state.Info()
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
	state := app.state.Append()

	if module == ModuleNameBase {
		if key == ChainKey {
			app.info.SetChainID(state, value)
			return "Success"
		}
		return fmt.Sprintf("Error: unknown base option: %s", key)
	}

	log, err := app.handler.SetOption(app.logger, state, module, key, value)
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

	ctx := stack.NewContext(
		app.GetChainID(),
		app.height,
		app.logger.With("call", "delivertx"),
	)
	res, err := app.handler.DeliverTx(ctx, app.state.Append(), tx)

	if err != nil {
		return errors.Result(err)
	}
	return res.ToABCI()
}

// CheckTx - ABCI
func (app *Basecoin) CheckTx(txBytes []byte) abci.Result {
	tx, err := basecoin.LoadTx(txBytes)
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
	return res.ToABCI()
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
