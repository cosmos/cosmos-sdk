package baseapp

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestApp wraps BaseApp with helper methods,
// and exposes more functionality than otherwise needed.
type TestApp struct {
	*BaseApp

	// These get set as we receive them.
	*abci.ResponseBeginBlock
	*abci.ResponseEndBlock
}

func NewTestApp(bapp *BaseApp) *TestApp {
	app := &TestApp{
		BaseApp: bapp,
	}
	return app
}

func (tapp *TestApp) RunBeginBlock() {
	if tapp.header != nil {
		panic("TestApp.header not nil, BeginBlock already run, or EndBlock not yet run.")
	}
	cms := tapp.CommitMultiStore()
	lastCommit := cms.LastCommitID()
	header := abci.Header{
		ChainID:        "chain_" + tapp.BaseApp.name,
		Height:         lastCommit.Version + 1,
		Time:           -1, // TODO
		NumTxs:         -1, // TODO
		LastCommitHash: lastCommit.Hash,
		DataHash:       nil, // TODO
		ValidatorsHash: nil, // TODO
		AppHash:        nil, // TODO
	}
	res := tapp.BeginBlock(abci.RequestBeginBlock{
		Hash:                nil, // TODO
		Header:              header,
		AbsentValidators:    nil, // TODO
		ByzantineValidators: nil, // TODO
	})
	tapp.ResponseBeginBlock = &res
	return
}

func (tapp *TestApp) ensureBeginBlock() {
	if tapp.header == nil {
		panic("TestApp.header was nil, call TestApp.RunBeginBlock()")
	}
}

func (tapp *TestApp) RunCheckTx(tx sdk.Tx) sdk.Result {
	tapp.ensureBeginBlock()
	return tapp.BaseApp.runTx(true, nil, tx)
}

func (tapp *TestApp) RunDeliverTx(tx sdk.Tx) sdk.Result {
	tapp.ensureBeginBlock()
	return tapp.BaseApp.runTx(true, nil, tx)
}

// NOTE: Skips authentication by wrapping msg in testTx{}.
func (tapp *TestApp) RunCheckMsg(msg sdk.Msg) sdk.Result {
	var tx = testTx{msg}
	return tapp.RunCheckTx(tx)
}

// NOTE: Skips authentication by wrapping msg in testTx{}.
func (tapp *TestApp) RunDeliverMsg(msg sdk.Msg) sdk.Result {
	var tx = testTx{msg}
	return tapp.RunCheckTx(tx)
}

func (tapp *TestApp) CommitMultiStore() sdk.CommitMultiStore {
	return tapp.BaseApp.cms
}

func (tapp *TestApp) MultiStoreCheck() sdk.MultiStore {
	return tapp.BaseApp.msCheck
}

func (tapp *TestApp) MultiStoreDeliver() sdk.MultiStore {
	return tapp.BaseApp.msDeliver
}

//----------------------------------------
// testTx

type testTx struct {
	sdk.Msg
}

func (tx testTx) GetSigners() []crypto.Address      { return nil }
func (tx testTx) GetFeePayer() crypto.Address       { return nil }
func (tx testTx) GetSignatures() []sdk.StdSignature { return nil }

func IsTestAppTx(tx sdk.Tx) bool {
	_, ok := tx.(testTx)
	return ok
}
