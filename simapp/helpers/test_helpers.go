package helpers

import (
	"math/rand"
	"time"

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
	fee := gen.NewFee()
	fee.SetAmount(feeAmt)
	fee.SetGas(gas)

	sigs := make([]context.ClientSignature, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	for i, p := range priv {
		// use a empty chainID for ease of testing
		clientSig := gen.NewSignature()
		clientSig.SetPubKey(p.PubKey())

		sigs[i] = clientSig
	}

	tx := gen.NewTx()
	tx.SetMsgs(msgs...)
	tx.SetFee(fee)
	tx.SetSignatures(sigs...)
	tx.SetMemo(memo)

	for i, p := range priv {
		// use a empty chainID for ease of testing
		signBytes, err := tx.CanonicalSignBytes(chainID, accnums[i], seq[i])
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i].SetSignature(sig)
	}

	tx.SetSignatures(sigs...)

	return tx.GetTx()
}
