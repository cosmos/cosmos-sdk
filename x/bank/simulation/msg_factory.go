package simulation

import (
	"math/rand"
	"testing"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// weight  (in context of the module)
// msg

// need:
// * account factory: existing account
// * banlance.SpendableCoins

// on failure: simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid transfers")
// on success: simtypes.NewOperationMsg(msg, true, "")

// helper
// random coins
// random fees

type SimAccount struct {
	simulation.Account
	r             *rand.Rand
	liquidBalance *SimsAccountBalance
	addressCodec  address.Codec
}

func (a SimAccount) LiquidBalance() *SimsAccountBalance {
	return a.liquidBalance
}

func (a SimAccount) AddressString() string {
	addr, err := a.addressCodec.BytesToString(a.Account.Address)
	if err != nil {
		panic(err.Error()) // todo: should be handled better
	}
	return addr
}

type SimsAccountBalance struct {
	sdk.Coins
	r *rand.Rand
}

// RandSubsetCoins return a random amount from the current balance. This can be empty.
// The amount is removed from the liquid balance.
func (b *SimsAccountBalance) RandSubsetCoins() sdk.Coins {
	amount := simtypes.RandSubsetCoins(b.r, b.Coins)
	b.Coins = b.Coins.Sub(amount...)
	return amount
}

func (b *SimsAccountBalance) RandFees() sdk.Coins {
	amount, err := simtypes.RandomFees(b.r, b.Coins)
	if err != nil {
		panic(err.Error()) // todo: setup a better way to abort execution
	}
	return amount
}

type SimAccountFilter func(a SimAccount) bool // returns true to accept

func NotEqualToAccount(other SimAccount) SimAccountFilter {
	return func(a SimAccount) bool {
		return !a.Address.Equals(other.Address)
	}
}

// todo: liquid token but sent may not be enabled for all or any
func WithSpendableBalance() SimAccountFilter {
	return func(a SimAccount) bool {
		return !a.liquidBalance.Empty()
	}
}

type ChainDataSource struct {
	t        *testing.T
	r        *rand.Rand
	accounts []SimAccount
}

func NewChainDataSource(t *testing.T, r *rand.Rand, oldSimAcc ...simulation.Account) *ChainDataSource {
	acc := make([]SimAccount, len(oldSimAcc))
	for i, a := range oldSimAcc {
		acc[i] = SimAccount{
			Account:       a,
			r:             r,
			liquidBalance: nil, // todo
		}
	}
	return &ChainDataSource{t: t, r: r}
}

func (c *ChainDataSource) AnyAccount(constrains ...SimAccountFilter) SimAccount {
	acc := c.randomAccount(10, constrains...)
	return acc
}

//

func (c *ChainDataSource) randomAccount(retryCount int, filters ...SimAccountFilter) SimAccount {
	if retryCount < 0 {
		c.t.Skip("failed to find a matching account")
		return SimAccount{}
	}
	idx := c.r.Intn(len(c.accounts))
	acc := c.accounts[idx]
	for _, constr := range filters {
		if !constr(acc) {
			return c.randomAccount(retryCount-1, filters...)
		}
	}
	return acc
}
