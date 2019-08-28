package helpers

import (
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenTx generates a signed mock transaction.
func GenTx(msgs []sdk.Msg, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) auth.StdTx {
	// Make the transaction free
	fee := auth.StdFee{
		Amount: sdk.NewCoins(sdk.NewInt64Coin("foocoin", 0)), // TODO: this should be the default bond denom
		Gas:    100000,
	}

	sigs := make([]auth.StdSignature, len(priv))

	// create a random length memo
	seed := rand.Int63()
	r := rand.New(rand.NewSource(seed))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 140))

	for i, p := range priv {
		// use a empty chainID for ease of testing
		sig, err := p.Sign(auth.StdSignBytes("", accnums[i], seq[i], fee, msgs, memo))
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    p.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, memo)
}
