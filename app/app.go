package app

import (
	"strings"

	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/governmint/gov"
	eyes "github.com/tendermint/merkleeyes/client"
	tmsp "github.com/tendermint/tmsp/types"
)

const (
	version   = "0.1"
	maxTxSize = 10240

	typeByteBase = 0x01
	typeByteGov  = 0x02

	pluginNameBase = "base"
	pluginNameGov  = "gov"
)

type Basecoin struct {
	eyesCli *eyes.Client
	govMint *gov.Governmint
	state   *state.State
	plugins *types.Plugins
}

func NewBasecoin(eyesCli *eyes.Client) *Basecoin {
	govMint := gov.NewGovernmint(eyesCli)
	state_ := state.NewState(eyesCli)
	plugins := types.NewPlugins()
	plugins.RegisterPlugin(typeByteGov, pluginNameGov, govMint) // TODO: make constants
	return &Basecoin{
		eyesCli: eyesCli,
		govMint: govMint,
		state:   state_,
		plugins: plugins,
	}
}

// TMSP::Info
func (app *Basecoin) Info() string {
	return Fmt("Basecoin v%v", version)
}

// TMSP::SetOption
func (app *Basecoin) SetOption(key string, value string) (log string) {
	pluginName, key := splitKey(key)
	if pluginName != pluginNameBase {
		// Set option on plugin
		plugin := app.plugins.GetByName(pluginName)
		if plugin == nil {
			return "Invalid plugin name: " + pluginName
		}
		return plugin.SetOption(key, value)
	} else {
		// Set option on basecoin
		switch key {
		case "chainID":
			app.state.SetChainID(value)
			return "Success"
		case "account":
			var err error
			var setAccount types.Account
			wire.ReadJSONPtr(&setAccount, []byte(value), &err)
			if err != nil {
				return "Error decoding setAccount message: " + err.Error()
			}
			accBytes := wire.BinaryBytes(setAccount)
			res := app.eyesCli.SetSync(setAccount.PubKey.Address(), accBytes)
			if res.IsErr() {
				return "Error saving account: " + res.Error()
			}
			return "Success"
		}
		return "Unrecognized option key " + key
	}
}

// TMSP::AppendTx
func (app *Basecoin) AppendTx(txBytes []byte) (res tmsp.Result) {
	if len(txBytes) > maxTxSize {
		return tmsp.ErrBaseEncodingError.AppendLog("Tx size exceeds maximum")
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return tmsp.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}
	// Validate and exec tx
	res = state.ExecTx(app.state, app.plugins, tx, false, nil)
	if res.IsErr() {
		return res.PrependLog("Error in AppendTx")
	}
	return tmsp.OK
}

// TMSP::CheckTx
func (app *Basecoin) CheckTx(txBytes []byte) (res tmsp.Result) {
	if len(txBytes) > maxTxSize {
		return tmsp.ErrBaseEncodingError.AppendLog("Tx size exceeds maximum")
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return tmsp.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}
	// Validate tx
	res = state.ExecTx(app.state, app.plugins, tx, true, nil)
	if res.IsErr() {
		return res.PrependLog("Error in CheckTx")
	}
	return tmsp.OK
}

// TMSP::Query
func (app *Basecoin) Query(query []byte) (res tmsp.Result) {
	pluginName, queryStr := splitKey(string(query))
	if pluginName != pluginNameBase {
		plugin := app.plugins.GetByName(pluginName)
		if plugin == nil {
			return tmsp.ErrBaseUnknownPlugin.SetLog(Fmt("Unknown plugin %v", pluginName))
		}
		return plugin.Query([]byte(queryStr))
	} else {
		// TODO turn Basecoin ops into a plugin?
		res = app.eyesCli.GetSync([]byte(queryStr))
		if res.IsErr() {
			return res.PrependLog("Error querying eyesCli")
		}
		return res
	}
}

// TMSP::Commit
func (app *Basecoin) Commit() (res tmsp.Result) {
	// First, commit all the plugins
	for _, plugin := range app.plugins.GetList() {
		res = plugin.Commit()
		if res.IsErr() {
			PanicSanity(Fmt("Error committing plugin %v", plugin.Name))
		}
	}
	// Then, commit eyes.
	res = app.eyesCli.CommitSync()
	if res.IsErr() {
		PanicSanity("Error getting hash: " + res.Error())
	}
	return res
}

// TMSP::InitChain
func (app *Basecoin) InitChain(validators []*tmsp.Validator) {
	app.govMint.InitChain(validators)
}

// TMSP::EndBlock
func (app *Basecoin) EndBlock(height uint64) []*tmsp.Validator {
	app.state.ResetCacheState()
	return app.govMint.EndBlock(height)
}

//----------------------------------------

// Splits the string at the first :.
// if there are none, the second string is nil.
func splitKey(key string) (prefix string, sufix string) {
	if strings.Contains(key, "/") {
		keyParts := strings.SplitN(key, "/", 2)
		return keyParts[0], keyParts[1]
	}
	return key, ""
}
