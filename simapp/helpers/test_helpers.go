package helpers

import (
	"errors"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenTx generates a signed mock transaction.
func GenTx(msgs []sdk.Msg, feeAmt sdk.Coins, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) auth.StdTx {
	fee := auth.StdFee{
		Amount: feeAmt,
		Gas:    100000, // TODO: this should be a param
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

// RandomFees generates a random fee amount for StdTx
func RandomFees(r *rand.Rand, ctx sdk.Context, acc authexported.Account,
	msgAmount sdk.Coins) (sdk.Coins, error) {
	// subtract the msg amount from the available coins
	coins, hasNeg := acc.GetCoins().SafeSub(msgAmount)
	if hasNeg {
		return nil, errors.New("not enough funds for transaction")
	}

	denomIndex := r.Intn(len(coins))
	randCoin := coins[denomIndex]

	amt, err := simulation.RandPositiveInt(r, randCoin.Amount)
	if err != nil {
		return nil, err
	}

	// Create a random fee and verify the fees are within the account's spendable
	// balance.
	fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))
	spendableCoins, hasNeg := acc.SpendableCoins(ctx.BlockHeader().Time).SafeSub(msgAmount)
	if hasNeg {
		return nil, errors.New("not enough funds for transaction")
	}

	if _, hasNeg = spendableCoins.SafeSub(fees); hasNeg {
		return nil, errors.New("not enough funds for fees")
	}

	return fees, nil
}
