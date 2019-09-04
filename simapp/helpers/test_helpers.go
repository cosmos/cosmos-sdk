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

// SimAppChainID hardcoded chainID for simulation
const SimAppChainID = "simulation-app"

// GenTx generates a signed mock transaction.
func GenTx(msgs []sdk.Msg, feeAmt sdk.Coins, chainID string, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) auth.StdTx {
	fee := auth.StdFee{
		Amount: feeAmt,
		Gas:    1000000, // TODO: this should be a param
	}

	sigs := make([]auth.StdSignature, len(priv))

	// create a random length memo
	seed := rand.Int63()
	r := rand.New(rand.NewSource(seed))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	for i, p := range priv {
		// use a empty chainID for ease of testing
		sig, err := p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, memo))
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

// RandomFees returns a random fee by selecting a random coin denomination and
// amount from the account's available balance. If the user doesn't have enough
// funds for paying fees, it returns empty coins.
func RandomFees(r *rand.Rand, ctx sdk.Context, acc authexported.Account,
	msgAmount sdk.Coins) (sdk.Coins, error) {
	if acc == nil {
		return nil, errors.New("account provided is nil")
	}

	// subtract the msg amount from the available coins
	coins := acc.SpendableCoins(ctx.BlockHeader().Time)
	coins, hasNeg := coins.SafeSub(msgAmount)
	if hasNeg {
		return nil, nil
	}

	if coins.Empty() {
		return nil, nil
	}

	denomIndex := r.Intn(len(coins))
	randCoin := coins[denomIndex]

	if randCoin.Amount.IsZero() {
		return nil, nil
	}

	amt, err := simulation.RandPositiveInt(r, randCoin.Amount)
	if err != nil {
		return nil, err
	}

	// Create a random fee and verify the fees are within the account's spendable
	// balance.
	fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))
	if _, hasNeg = coins.SafeSub(fees); hasNeg {
		return nil, errors.New("not enough funds for fees")
	}

	return fees, nil
}
