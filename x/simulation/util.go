package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func getTestingMode(tb testing.TB) (testingMode bool, t *testing.T, b *testing.B) {
	testingMode = false

	if _t, ok := tb.(*testing.T); ok {
		t = _t
		testingMode = true
	} else {
		b = tb.(*testing.B)
	}

	return testingMode, t, b
}

// getBlockSize returns a block size as determined from the transition matrix.
// It targets making average block size the provided parameter. The three
// states it moves between are:
//  - "over stuffed" blocks with average size of 2 * avgblocksize,
//  - normal sized blocks, hitting avgBlocksize on average,
//  - and empty blocks, with no txs / only txs scheduled from the past.
func getBlockSize(r *rand.Rand, params Params, lastBlockSizeState, avgBlockSize int) (state, blockSize int) {
	// TODO: Make default blocksize transition matrix actually make the average
	// blocksize equal to avgBlockSize.
	state = params.BlockSizeTransitionMatrix().NextState(r, lastBlockSizeState)

	switch state {
	case 0:
		blockSize = r.Intn(avgBlockSize * 4)

	case 1:
		blockSize = r.Intn(avgBlockSize * 2)

	default:
		blockSize = 0
	}

	return state, blockSize
}

func mustMarshalJSONIndent(o interface{}) []byte {
	bz, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("failed to JSON encode: %s", err))
	}

	return bz
}

// GenAndDeliverTxWithRandFees generates a transaction with a random fee and delivers it.
func GenAndDeliverTxWithRandFees(
	r *rand.Rand,
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	cdc *codec.ProtoCodec,
	msg sdk.Msg,
	msgType string,
	coinsSpentInMsg sdk.Coins,
	ctx sdk.Context,
	simAccount simtypes.Account,
	ak AccountKeeper,
	bk BankKeeper,
	moduleName string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(coinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(moduleName, msgType, "message doesn't leave room for fees"), nil, err
	}

	fees, err = simtypes.RandomFees(r, ctx, coins)
	if err != nil {
		return simtypes.NoOpMsg(moduleName, msgType, "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(app, txGen, cdc, msg, msgType, fees, ctx, simAccount, ak, moduleName)
}

// GenAndDeliverTx generates a transactions and delivers it.
func GenAndDeliverTx(
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	cdc *codec.ProtoCodec,
	msg sdk.Msg,
	msgType string,
	fees sdk.Coins,
	ctx sdk.Context,
	simAccount simtypes.Account,
	ak AccountKeeper,
	moduleName string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)

	if err != nil {
		return simtypes.NoOpMsg(moduleName, msgType, "unable to generate mock tx"), nil, err
	}

	_, _, err = app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(moduleName, msgType, "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(msg, true, "", cdc), nil, nil

}
