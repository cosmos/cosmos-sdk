package mock

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// CheckBalance checks the balance of an account.
func CheckBalance(t *testing.T, app *App, addr sdk.AccAddress, exp sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	res := app.AccountMapper.GetAccount(ctxCheck, addr)

	require.Equal(t, exp, res.GetCoins())
}

// CheckGenTx checks a generated signed transaction. The result of the check is
// compared against the parameter 'expPass'. A test assertion is made using the
// parameter 'expPass' against the result. A corresponding result is returned.
func CheckGenTx(
	t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg, accNums []int64,
	seq []int64, expPass bool, priv ...crypto.PrivKey,
) sdk.Result {
	tx := GenTx(msgs, accNums, seq, priv...)
	res := app.Check(tx)

	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	return res
}

// SignCheckDeliver checks a generated signed transaction and simulates a
// block commitment with the given transaction. A test assertion is made using
// the parameter 'expPass' against the result. A corresponding result is
// returned.
func SignCheckDeliver(
	t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg, accNums []int64,
	seq []int64, expPass bool, priv ...crypto.PrivKey,
) sdk.Result {
	tx := GenTx(msgs, accNums, seq, priv...)
	res := app.Check(tx)

	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	// Simulate a sending a transaction and committing a block
	app.BeginBlock(abci.RequestBeginBlock{})
	res = app.Deliver(tx)

	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return res
}
