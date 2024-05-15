package simulation

import (
	"context"
	"math/rand"
	"slices"

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

// useful helpers:
// RandSubsetCoins

type SimAccount struct {
	simulation.Account
	r             *rand.Rand
	liquidBalance *SimsAccountBalance
	addressCodec  address.Codec
	bank          SpendableCoinser
}

func (a *SimAccount) LiquidBalance() *SimsAccountBalance {
	if a.liquidBalance == nil {
		a.liquidBalance = NewSimsAccountBalance(a.r, a.bank.SpendableCoins(a.Address))
	}
	return a.liquidBalance
}

func (a SimAccount) AddressString() string {
	addr, err := a.addressCodec.BytesToString(a.Account.Address)
	if err != nil {
		panic(err.Error()) // todo: should be handled better
	} // todo: what are the scenarios that this can fail?
	return addr
}

type SimsAccountBalance struct {
	sdk.Coins
	r *rand.Rand
}

func NewSimsAccountBalance(r *rand.Rand, coins sdk.Coins) *SimsAccountBalance {
	return &SimsAccountBalance{Coins: coins, r: r}
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
	} // todo: revisit the panic
	return amount
}

type SimAccountFilter func(a SimAccount) bool // returns true to accept

func ExcludeAccounts(others ...SimAccount) SimAccountFilter {
	return func(a SimAccount) bool {
		return slices.ContainsFunc(others, func(o SimAccount) bool {
			return !a.Address.Equals(o.Address)
		})
	}
}

func ExcludeAddresses(addrs ...string) SimAccountFilter {
	return func(a SimAccount) bool {
		return !slices.Contains(addrs, a.AddressString())
	}
}

// todo: liquid token but sent may not be enabled for all or any
func WithSpendableBalance() SimAccountFilter {
	return func(a SimAccount) bool {
		return !a.LiquidBalance().Empty()
	}
}

type ModuleAccountSource interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}
type SpendableCoinser interface {
	SpendableCoins(addr sdk.AccAddress) sdk.Coins
}

type SpendableCoinserFn func(addr sdk.AccAddress) sdk.Coins

func (b SpendableCoinserFn) SpendableCoins(addr sdk.AccAddress) sdk.Coins {
	return b(addr)
}

type BalanceSource interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

func NewBalanceSource(ctx sdk.Context, bk BalanceSource,
) SpendableCoinser {
	return SpendableCoinserFn(func(addr sdk.AccAddress) sdk.Coins {
		return bk.SpendableCoins(ctx, addr)
	})
}

type ChainDataSource struct {
	r             *rand.Rand
	accounts      []SimAccount
	accountSource ModuleAccountSource
	addressCodec  address.Codec
}

func NewChainDataSource(r *rand.Rand, ak ModuleAccountSource, bk SpendableCoinser, codec address.Codec, oldSimAcc ...simtypes.Account) *ChainDataSource {
	acc := make([]SimAccount, len(oldSimAcc))
	for i, a := range oldSimAcc {
		acc[i] = SimAccount{
			Account:      a,
			r:            r,
			addressCodec: codec,
			bank:         bk,
		}
	}
	return &ChainDataSource{r: r, accountSource: ak, addressCodec: codec, accounts: acc}
}

// not module account
func (c *ChainDataSource) AnyAccount(r SimulationReporter, constrains ...SimAccountFilter) SimAccount {
	acc := c.randomAccount(r, 10, constrains...)
	return acc
}

//

func (c *ChainDataSource) randomAccount(reporter SimulationReporter, retryCount int, filters ...SimAccountFilter) SimAccount {
	if retryCount < 0 {
		reporter.Skip("failed to find a matching account")
		return SimAccount{}
	}
	idx := c.r.Intn(len(c.accounts))
	acc := c.accounts[idx]
	for _, constr := range filters {
		if !constr(acc) {
			return c.randomAccount(nil, retryCount-1, filters...)
		}
	}
	return acc
}

func (c *ChainDataSource) ModuleAccountAddress(reporter SimulationReporter, moduleName string) string {
	acc := c.accountSource.GetModuleAddress(moduleName)
	if acc == nil {
		reporter.Skipf("unknown module account: %s", moduleName)
		return ""
	}
	res, err := c.addressCodec.BytesToString(acc)
	if err != nil {
		reporter.Skipf("failed to encode module address: %s", err)
		return ""
	}
	return res
}

func (c *ChainDataSource) Rand() *rand.Rand {
	return c.r
}

// Collect applies the function f to each element in the source slice,
// returning a new slice containing the results.
//
// The source slice can contain elements of any type T, and the function f
// should take an element of type T as input and return a value of any type E.
//
// Example usage:
//
//	source := []int{1, 2, 3, 4, 5}
//	double := Collect(source, func(x int) int {
//	    return x * 2
//	})
//	// double is now []int{2, 4, 6, 8, 10}
func Collect[T, E any](source []T, f func(a T) E) []E {
	r := make([]E, len(source))
	for i, v := range source {
		r[i] = f(v)
	}
	return r
}
