package mock

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BigInterval is a representation of the interval [lo, hi), where
// lo and hi are both of type sdk.Int
type BigInterval struct {
	lo sdk.Int
	hi sdk.Int
}

// RandFromBigInterval chooses an interval uniformly from the provided list of
// BigIntervals, and then chooses an element from an interval uniformly at random.
func RandFromBigInterval(r *rand.Rand, intervals []BigInterval) sdk.Int {
	if len(intervals) == 0 {
		return sdk.ZeroInt()
	}

	interval := intervals[r.Intn(len(intervals))]

	lo := interval.lo
	hi := interval.hi

	diff := hi.Sub(lo)
	result := sdk.NewIntFromBigInt(new(big.Int).Rand(r, diff.BigInt()))
	result = result.Add(lo)

	return result
}

// CheckBalance checks the balance of an account.
func CheckBalance(t *testing.T, app *App, addr sdk.AccAddress, exp sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	res := app.AccountKeeper.GetAccount(ctxCheck, addr)

	require.Equal(t, exp, res.GetCoins())
}

// CheckGenTx checks a generated signed transaction. The result of the check is
// compared against the parameter 'expPass'. A test assertion is made using the
// parameter 'expPass' against the result. A corresponding result is returned.
func CheckGenTx(
	t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg, accNums []uint64,
	seq []uint64, expPass bool, priv ...crypto.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx := GenTx(msgs, accNums, seq, priv...)
	gInfo, res, err := app.Check(tx)

	if expPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	return gInfo, res, err
}

// SignCheckDeliver checks a generated signed transaction and simulates a
// block commitment with the given transaction. A test assertion is made using
// the parameter 'expPass' against the result. A corresponding result is
// returned.
func SignCheckDeliver(
	t *testing.T, cdc *codec.Codec, app *baseapp.BaseApp, header abci.Header, msgs []sdk.Msg,
	accNums, seq []uint64, expSimPass, expPass bool, priv ...crypto.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {

	tx := GenTx(msgs, accNums, seq, priv...)

	txBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)
	require.Nil(t, err)

	// Must simulate now as CheckTx doesn't run Msgs anymore
	_, res, err := app.Simulate(txBytes, tx, 0)

	if expSimPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	// Simulate a sending a transaction and committing a block
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	gInfo, res, err := app.Deliver(tx)

	if expPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return gInfo, res, err
}
