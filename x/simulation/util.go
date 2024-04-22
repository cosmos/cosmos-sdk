package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	authsign "cosmossdk.io/x/auth/signing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

const DefaultGenTxGas = 10000000

func getTestingMode(tb testing.TB) (testingMode bool, t *testing.T, b *testing.B) {
	tb.Helper()
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
//   - "over stuffed" blocks with average size of 2 * avgblocksize,
//   - normal sized blocks, hitting avgBlocksize on average,
//   - and empty blocks, with no txs / only txs scheduled from the past.
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

// OperationInput is a struct that holds all the needed values to generate a tx and deliver it
type OperationInput struct {
	R               *rand.Rand
	App             *baseapp.BaseApp
	TxGen           client.TxConfig
	Cdc             *codec.ProtoCodec
	Msg             sdk.Msg
	CoinsSpentInMsg sdk.Coins
	Context         sdk.Context
	SimAccount      simtypes.Account
	AccountKeeper   AccountKeeper
	Bankkeeper      BankKeeper
	ModuleName      string
}

// GenAndDeliverTxWithRandFees generates a transaction with a random fee and delivers it.
func GenAndDeliverTxWithRandFees(txCtx OperationInput) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	spendable := txCtx.Bankkeeper.SpendableCoins(txCtx.Context, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(txCtx.CoinsSpentInMsg...)
	if hasNeg {
		return simtypes.NoOpMsg(txCtx.ModuleName, sdk.MsgTypeURL(txCtx.Msg), "message doesn't leave room for fees"), nil, err
	}

	fees, err = simtypes.RandomFees(txCtx.R, coins)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, sdk.MsgTypeURL(txCtx.Msg), "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(txCtx, fees)
}

// GenAndDeliverTx generates a transactions and delivers it.
func GenAndDeliverTx(txCtx OperationInput, fees sdk.Coins) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	tx, err := GenSignedMockTx(
		txCtx.R,
		txCtx.TxGen,
		[]sdk.Msg{txCtx.Msg},
		fees,
		10000000, //Default gen tx gas
		txCtx.Context.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		txCtx.SimAccount.PrivKey,
	)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, sdk.MsgTypeURL(txCtx.Msg), "unable to generate mock tx"), nil, err
	}

	_, _, err = txCtx.App.SimDeliver(txCtx.TxGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, sdk.MsgTypeURL(txCtx.Msg), "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(txCtx.Msg, true, ""), nil, nil
}

// GenSignedMockTx generates a signed mock transaction.
func GenSignedMockTx(r *rand.Rand, txConfig client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode, err := authsign.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	tx := txConfig.NewTxBuilder()
	err = tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			Address:       sdk.AccAddress(p.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        p.PubKey(),
		}

		signBytes, err := authsign.GetSignBytesAdapter(
			context.Background(), txConfig.SignModeHandler(), signMode, signerData,
			tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}
