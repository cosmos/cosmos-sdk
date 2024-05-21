package simsx

import (
	"context"
	"math/rand"
	"slices"
	"time"

	"cosmossdk.io/math"

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
	AddressBech32 string
	bank          SpendableCoinser
}

func (a *SimAccount) LiquidBalance() *SimsAccountBalance {
	if a.liquidBalance == nil {
		a.liquidBalance = NewSimsAccountBalance(a.r, a.bank.SpendableCoins(a.Address))
	}
	return a.liquidBalance
}

type SimsAccountBalance struct {
	sdk.Coins
	r *rand.Rand
}

func NewSimsAccountBalance(r *rand.Rand, coins sdk.Coins) *SimsAccountBalance {
	return &SimsAccountBalance{Coins: coins, r: r}
}

type CoinsFilter func(c sdk.Coins) bool // returns false to reject

func WithSendEnabledCoins(ctx context.Context, bk interface {
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
},
) CoinsFilter {
	return func(coins sdk.Coins) bool {
		return bk.IsSendEnabledCoins(ctx, coins...) == nil
	}
}

// RandSubsetCoins return random amounts from the current balance. When the coins are empty, skip is called on the reporter.
// The amounts are removed from the liquid balance.
func (b *SimsAccountBalance) RandSubsetCoins(reporter SimulationReporter, filters ...CoinsFilter) sdk.Coins {
	amount := b.randomAmount(5, reporter, b.Coins, filters...)

	b.Coins = b.Coins.Sub(amount...)
	if !amount.Empty() {
		reporter.Skip("got empty amounts")
	}
	return amount
}

func (b *SimsAccountBalance) RandSubsetCoin(reporter SimulationReporter, denom string, filters ...CoinsFilter) sdk.Coin {
	ok, coin := b.Find(denom)
	if !ok {
		reporter.Skipf("no such coin: %s", denom)
		return sdk.Coin{}
	}
	amount := b.randomAmount(5, reporter, sdk.NewCoins(coin), filters...)
	b.Coins = b.Coins.Sub(amount...)
	_, coin = amount.Find(denom)
	if !coin.IsPositive() {
		reporter.Skip("empty coin")
	}
	return coin
}

func (b *SimsAccountBalance) randomAmount(retryCount int, reporter SimulationReporter, coins sdk.Coins, filters ...CoinsFilter) sdk.Coins {
	if retryCount < 0 {
		reporter.Skip("failed to find matching amount")
		return sdk.Coins{}
	}
	amount := simtypes.RandSubsetCoins(b.r, coins)
	for _, filter := range filters {
		if !filter(amount) {
			return b.randomAmount(retryCount-1, reporter, coins, filters...)
		}
	}
	return amount
}

func (b *SimsAccountBalance) RandFees() sdk.Coins {
	amount, err := simtypes.RandomFees(b.r, b.Coins)
	if err != nil {
		panic(err.Error()) // todo: setup a better way to abort execution
	} // todo: revisit the panic
	return amount
}

type SimAccountFilter func(a SimAccount) bool // returns false to reject

func ExcludeAccounts(others ...SimAccount) SimAccountFilter {
	return func(a SimAccount) bool {
		return !slices.ContainsFunc(others, func(o SimAccount) bool {
			return a.Address.Equals(o.Address)
		})
	}
}

func ExcludeAddresses(addrs ...string) SimAccountFilter {
	return func(a SimAccount) bool {
		return !slices.Contains(addrs, a.AddressBech32)
	}
}

func WithDenomBalance(denom string) SimAccountFilter {
	return func(a SimAccount) bool {
		return a.LiquidBalance().AmountOf(denom).IsPositive()
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
	r                         *rand.Rand
	addressToAccountsPosIndex map[string]int
	accounts                  []SimAccount
	accountSource             ModuleAccountSource
	addressCodec              address.Codec
}

func NewChainDataSource(r *rand.Rand, ak ModuleAccountSource, bk SpendableCoinser, codec address.Codec, oldSimAcc ...simtypes.Account) *ChainDataSource {
	acc := make([]SimAccount, len(oldSimAcc))
	index := make(map[string]int, len(oldSimAcc))
	for i, a := range oldSimAcc {
		addrStr, err := codec.BytesToString(a.Address)
		if err != nil {
			panic(err.Error())
		}
		acc[i] = SimAccount{
			Account:       a,
			r:             r,
			AddressBech32: addrStr,
			bank:          bk,
		}
		index[addrStr] = i
	}
	return &ChainDataSource{r: r, accountSource: ak, addressCodec: codec, accounts: acc}
}

// no module accounts
func (c *ChainDataSource) AnyAccount(r SimulationReporter, filters ...SimAccountFilter) SimAccount {
	acc := c.randomAccount(r, 5, filters...)
	return acc
}

func (c ChainDataSource) GetAccount(reporter SimulationReporter, addr string) SimAccount {
	pos, ok := c.addressToAccountsPosIndex[addr]
	if !ok {
		reporter.Skipf("no such account: %s", addr)
		return SimAccount{}
	}
	return c.accounts[pos]
}

func (c *ChainDataSource) randomAccount(reporter SimulationReporter, retryCount int, filters ...SimAccountFilter) SimAccount {
	if retryCount < 0 {
		reporter.Skip("failed to find a matching account")
		return SimAccount{}
	}
	idx := c.r.Intn(len(c.accounts))
	acc := c.accounts[idx]
	for _, filter := range filters {
		if !filter(acc) {
			return c.randomAccount(reporter, retryCount-1, filters...)
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

func (c *ChainDataSource) Rand() *XRand {
	return &XRand{c.r}
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

type XRand struct {
	*rand.Rand
}

func NewXRand(rand *rand.Rand) *XRand {
	return &XRand{Rand: rand}
}

func (r *XRand) StringN(max int) string {
	return simtypes.RandStringOfLength(r.Rand, max)
}

func (r *XRand) SubsetCoins(src sdk.Coins) sdk.Coins {
	return simulation.RandSubsetCoins(r.Rand, src)
}

func (r *XRand) DecN(max math.LegacyDec) math.LegacyDec {
	return simtypes.RandomDecAmount(r.Rand, max)
}

func (r *XRand) IntInRange(min, max int) int {
	return r.Rand.Intn(max-min) + min
}

func (r *XRand) PositiveInt(max math.Int) (math.Int, error) {
	return simtypes.RandPositiveInt(r.Rand, max)
}

// Timestamp returns a timestamp between  Jan 1, 2062 and Jan 1, 2262
func (r *XRand) Timestamp() time.Time {
	return simtypes.RandTimestamp(r.Rand)
}

func (r *XRand) Bool() bool {
	return r.Intn(100) > 50
}
