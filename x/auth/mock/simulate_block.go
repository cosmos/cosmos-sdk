package mock

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"

	abci "github.com/tendermint/abci/types"
)

var chainID = "" // TODO

// set the mock app genesis
func SetGenesis(app *App, accs []auth.Account) {

	// pass the accounts in via the application (lazy) instead of through RequestInitChain
	app.GenesisAccounts = accs

	app.InitChain(abci.RequestInitChain{})
	app.Commit()
}

// check an account balance
func CheckBalance(t *testing.T, app *App, addr sdk.Address, exp sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	res := app.AccountMapper.GetAccount(ctxCheck, addr)
	assert.Equal(t, exp, res.GetCoins())
}

// generate a signed transaction
func GenTx(msgs []sdk.Msg, accnums []int64, seq []int64, priv ...crypto.PrivKeyEd25519) auth.StdTx {

	// make the transaction free
	fee := auth.StdFee{
		sdk.Coins{sdk.NewCoin("foocoin", 0)},
		100000,
	}

	sigs := make([]auth.StdSignature, len(priv))
	memo := "testmemotestmemo"
	for i, p := range priv {
		sigs[i] = auth.StdSignature{
			PubKey:        p.PubKey(),
			Signature:     p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, memo)),
			AccountNumber: accnums[i],
			Sequence:      seq[i],
		}
	}
	return auth.NewStdTx(msgs, fee, sigs, memo)
}

// check a transaction result
func SignCheck(t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg, accnums []int64, seq []int64, priv ...crypto.PrivKeyEd25519) sdk.Result {
	tx := GenTx(msgs, accnums, seq, priv...)
	res := app.Check(tx)
	return res
}

// simulate a block
func SignCheckDeliver(t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg, accnums []int64, seq []int64, expPass bool, priv ...crypto.PrivKeyEd25519) {

	// Sign the tx
	tx := GenTx(msgs, accnums, seq, priv...)

	// Run a Check
	res := app.Check(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	// Simulate a Block
	app.BeginBlock(abci.RequestBeginBlock{})
	res = app.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}
	app.EndBlock(abci.RequestEndBlock{})

	app.Commit()
}
