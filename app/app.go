package app

import (
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/governmint/gov"
	eyes "github.com/tendermint/merkleeyes/client"
	tmsp "github.com/tendermint/tmsp/types"
)

const version = "0.1"
const maxTxSize = 10240

type Basecoin struct {
	eyesCli *eyes.Client
	govMint *gov.Governmint
	state   *state.State
}

func NewBasecoin(eyesCli *eyes.Client) *Basecoin {
	state_ := state.NewState(eyesCli)
	govMint := gov.NewGovernmint(eyesCli)
	state_.RegisterPlugin([]byte("gov"), govMint)
	return &Basecoin{
		eyesCli: eyesCli,
		govMint: govMint,
		state:   state_,
	}
}

// TMSP::Info
func (app *Basecoin) Info() string {
	return Fmt("Basecoin v%v", version)
}

// TMSP::SetOption
func (app *Basecoin) SetOption(key string, value string) (log string) {
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
	res = state.ExecTx(app.state, tx, false, nil)
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
	res = state.ExecTx(app.state, tx, true, nil)
	if res.IsErr() {
		return res.PrependLog("Error in CheckTx")
	}
	return tmsp.OK
}

// TMSP::Query
func (app *Basecoin) Query(query []byte) (res tmsp.Result) {
	res = app.eyesCli.GetSync(query)
	if res.IsErr() {
		return res.PrependLog("Error querying eyesCli")
	}
	return res
}

// TMSP::Commit
func (app *Basecoin) Commit() (res tmsp.Result) {
	res = app.eyesCli.CommitSync()
	if res.IsErr() {
		panic("Error getting hash: " + res.Error())
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
