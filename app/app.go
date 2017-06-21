package app

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	sm "github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/basecoin/version"
)

const (
	maxTxSize      = 10240
	PluginNameBase = "base"
)

type Basecoin struct {
	eyesCli    *eyes.Client
	state      *sm.State
	cacheState *sm.State
	plugins    *types.Plugins
	logger     log.Logger
}

func NewBasecoin(eyesCli *eyes.Client) *Basecoin {
	state := sm.NewState(eyesCli)
	plugins := types.NewPlugins()
	return &Basecoin{
		eyesCli:    eyesCli,
		state:      state,
		cacheState: nil,
		plugins:    plugins,
		logger:     log.NewNopLogger(),
	}
}

func (app *Basecoin) SetLogger(l log.Logger) {
	app.logger = l
	app.state.SetLogger(l.With("module", "state"))
}

// XXX For testing, not thread safe!
func (app *Basecoin) GetState() *sm.State {
	return app.state.CacheWrap()
}

// ABCI::Info
func (app *Basecoin) Info() abci.ResponseInfo {
	resp, err := app.eyesCli.InfoSync()
	if err != nil {
		cmn.PanicCrisis(err)
	}
	return abci.ResponseInfo{
		Data:             cmn.Fmt("Basecoin v%v", version.Version),
		LastBlockHeight:  resp.LastBlockHeight,
		LastBlockAppHash: resp.LastBlockAppHash,
	}
}

func (app *Basecoin) RegisterPlugin(plugin types.Plugin) {
	app.plugins.RegisterPlugin(plugin)
}

// ABCI::SetOption
func (app *Basecoin) SetOption(key string, value string) string {
	pluginName, key := splitKey(key)
	if pluginName != PluginNameBase {
		// Set option on plugin
		plugin := app.plugins.GetByName(pluginName)
		if plugin == nil {
			return "Invalid plugin name: " + pluginName
		}
		app.logger.Info("SetOption on plugin", "plugin", pluginName, "key", key, "value", value)
		return plugin.SetOption(app.state, key, value)
	} else {
		// Set option on basecoin
		switch key {
		case "chain_id":
			app.state.SetChainID(value)
			return "Success"
		case "account":
			var acc GenesisAccount
			err := json.Unmarshal([]byte(value), &acc)
			if err != nil {
				return "Error decoding acc message: " + err.Error()
			}
			acc.Balance.Sort()
			addr, err := acc.GetAddr()
			if err != nil {
				return "Invalid address: " + err.Error()
			}
			app.state.SetAccount(addr, acc.ToAccount())
			app.logger.Info("SetAccount", "addr", hex.EncodeToString(addr), "acc", acc)

			return "Success"
		}
		return "Unrecognized option key " + key
	}
}

// ABCI::DeliverTx
func (app *Basecoin) DeliverTx(txBytes []byte) (res abci.Result) {
	if len(txBytes) > maxTxSize {
		return abci.ErrBaseEncodingError.AppendLog("Tx size exceeds maximum")
	}

	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Validate and exec tx
	res = sm.ExecTx(app.state, app.plugins, tx, false, nil)
	if res.IsErr() {
		return res.PrependLog("Error in DeliverTx")
	}
	return res
}

// ABCI::CheckTx
func (app *Basecoin) CheckTx(txBytes []byte) (res abci.Result) {
	if len(txBytes) > maxTxSize {
		return abci.ErrBaseEncodingError.AppendLog("Tx size exceeds maximum")
	}

	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Validate tx
	res = sm.ExecTx(app.cacheState, app.plugins, tx, true, nil)
	if res.IsErr() {
		return res.PrependLog("Error in CheckTx")
	}
	return abci.OK
}

// ABCI::Query
func (app *Basecoin) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if len(reqQuery.Data) == 0 {
		resQuery.Log = "Query cannot be zero length"
		resQuery.Code = abci.CodeType_EncodingError
		return
	}

	// handle special path for account info
	if reqQuery.Path == "/account" {
		reqQuery.Path = "/key"
		reqQuery.Data = types.AccountKey(reqQuery.Data)
	}

	resQuery, err := app.eyesCli.QuerySync(reqQuery)
	if err != nil {
		resQuery.Log = "Failed to query MerkleEyes: " + err.Error()
		resQuery.Code = abci.CodeType_InternalError
		return
	}
	return
}

// ABCI::Commit
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

// ABCI::InitChain
func (app *Basecoin) InitChain(validators []*abci.Validator) {
	for _, plugin := range app.plugins.GetList() {
		plugin.InitChain(app.state, validators)
	}
}

// ABCI::BeginBlock
func (app *Basecoin) BeginBlock(hash []byte, header *abci.Header) {
	for _, plugin := range app.plugins.GetList() {
		plugin.BeginBlock(app.state, hash, header)
	}
}

// ABCI::EndBlock
func (app *Basecoin) EndBlock(height uint64) (res abci.ResponseEndBlock) {
	for _, plugin := range app.plugins.GetList() {
		pluginRes := plugin.EndBlock(app.state, height)
		res.Diffs = append(res.Diffs, pluginRes.Diffs...)
	}
	return
}

//----------------------------------------

// Splits the string at the first '/'.
// if there are none, the second string is nil.
func splitKey(key string) (prefix string, suffix string) {
	if strings.Contains(key, "/") {
		keyParts := strings.SplitN(key, "/", 2)
		return keyParts[0], keyParts[1]
	}
	return key, ""
}
