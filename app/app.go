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
	govMint := gov.NewGovernmint(eyesCli)
	return &Basecoin{
		eyesCli: eyesCli,
		govMint: govMint,
		state:   state.NewState(eyesCli),
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
	case "account":
		var err error
		var setAccount types.Account
		wire.ReadJSONPtr(&setAccount, []byte(value), &err)
		if err != nil {
			return "Error decoding setAccount message: " + err.Error()
		}
		accBytes := wire.BinaryBytes(setAccount)
		err = app.eyesCli.SetSync(setAccount.PubKey.Address(), accBytes)
		if err != nil {
			return "Error saving account: " + err.Error()
		}
		return "Success"
	}
	return "Unrecognized option key " + key
}

// TMSP::AppendTx
func (app *Basecoin) AppendTx(txBytes []byte) (res tmsp.Result) {
	if len(txBytes) > maxTxSize {
		return types.ErrEncodingError.AppendLog("Tx size exceeds maximum")
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return types.ErrEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}
	// Validate and exec tx
	res = state.ExecTx(app.state, tx, false, nil)
	if !res.IsOK() {
		return res.PrependLog("Error in AppendTx")
	}
	return types.ResultOK
}

// TMSP::CheckTx
func (app *Basecoin) CheckTx(txBytes []byte) (res tmsp.Result) {
	if len(txBytes) > maxTxSize {
		return types.ErrEncodingError.AppendLog("Tx size exceeds maximum")
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return types.ErrEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}
	// Validate tx
	res = state.ExecTx(app.state, tx, true, nil)
	if !res.IsOK() {
		return res.PrependLog("Error in CheckTx")
	}
	return types.ResultOK
}

// TMSP::Query
func (app *Basecoin) Query(query []byte) (res tmsp.Result) {
	return types.ResultOK
	value, err := app.eyesCli.GetSync(query)
	if err != nil {
		panic("Error making query: " + err.Error())
	}
	return types.ResultOK.SetData(value).SetLog("Success")
}

// TMSP::Commit
func (app *Basecoin) Commit() (hash []byte, log string) {
	hash, log, err := app.eyesCli.CommitSync()
	if err != nil {
		panic("Error getting hash: " + err.Error())
	}
	return hash, "Success"
}

// TMSP::InitChain
func (app *Basecoin) InitChain(validators []*tmsp.Validator) {
	app.govMint.InitChain(validators)
}

// TMSP::EndBlock
func (app *Basecoin) EndBlock(height uint64) []*tmsp.Validator {
	return app.govMint.EndBlock(height)
}
