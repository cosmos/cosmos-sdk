package helpers

import (
	"math/rand"
	"time"

	types "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// SimAppChainID hardcoded chainID for simulation
const (
	DefaultGenTxGas = 1000000
	SimAppChainID   = "simulation-app"
)

// GenTx generates a signed mock transaction.
func GenTx(gen context.TxGenerator, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) sdk.Tx {
	sigs := make([]context.SignatureBuilder, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode := gen.SignModeHandler().DefaultMode()

	for i, p := range priv {
		sigs[i] = context.SignatureBuilder{
			PubKey: p.PubKey(),
			Data: &types.SingleSignatureData{
				SignMode: signMode,
			},
		}
	}

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		panic(err)
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		panic(err)
	}
	tx.SetMemo(memo)
	tx.SetFee(feeAmt)
	tx.SetGasLimit(gas)

	for i, p := range priv {
		// use a empty chainID for ease of testing
		signBytes, err := gen.SignModeHandler().GetSignBytes(types.SigningData{
			Mode:            0,
			PublicKey:       nil,
			ChainID:         chainID,
			AccountNumber:   accnums[i],
			AccountSequence: seq[i],
		}, tx.GetTx())

		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i].Data.(*types.SingleSignatureData).Signature = sig
	}

	err = tx.SetSignatures(sigs...)
	if err != nil {
		panic(err)
	}

	return tx.GetTx()
}
