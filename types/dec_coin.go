package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
// Decimal Coin

// DecCoin defines a coin which can have additional decimal points
type DecCoin struct {
	Denom  string `json:"denom"`
	Amount Dec    `json:"amount"`
}

// NewDecCoin creates a new DecCoin instance from an Int.
func NewDecCoin(denom string, amount Int) DecCoin {
	if err := validate(denom, amount); err != nil {
		panic(err)
	}

	return DecCoin{
		Denom:  denom,
		Amount: amount.ToDec(),
	}
}

// NewDecCoinFromDec creates a new DecCoin instance from a Dec.
func NewDecCoinFromDec(denom string, amount Dec) DecCoin {
	mustValidateDenom(denom)

	if amount.LT(ZeroDec()) {
		panic(fmt.Sprintf("negative decimal coin amount: %v\n", amount))
	}

	return DecCoin{
		Denom:  denom,
		Amount: amount,
	}
}

// NewDecCoinFromCoin creates a new DecCoin from a Coin.
func NewDecCoinFromCoin(coin Coin) DecCoin {
	if err := validate(coin.Denom, coin.Amount); err != nil {
		panic(err)
	}

	return DecCoin{
		Denom:  coin.Denom,
		Amount: coin.Amount.ToDec(),
	}
}

// NewInt64DecCoin returns a new DecCoin with a denomination and amount. It will
// panic if the amount is negative or denom is invalid.
func NewInt64DecCoin(denom string, amount int64) DecCoin {
	return NewDecCoin(denom, NewInt(amount))
}

// IsZero returns if the DecCoin amount is zero.
func (coin DecCoin) IsZero() bool {
	return coin.Amount.IsZero()
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value.
func (coin DecCoin) IsGTE(other DecCoin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}

	return !coin.Amount.LT(other.Amount)
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value.
func (coin DecCoin) IsLT(other DecCoin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}

	return coin.Amount.LT(other.Amount)
}

// IsEqual returns true if the two sets of Coins have the same value.
func (coin DecCoin) IsEqual(other DecCoin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}

	return coin.Amount.Equal(other.Amount)
}

// Add adds amounts of two decimal coins with same denom.
func (coin DecCoin) Add(coinB DecCoin) DecCoin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("coin denom different: %v %v\n", coin.Denom, coinB.Denom))
	}
	return DecCoin{coin.Denom, coin.Amount.Add(coinB.Amount)}
}

// Sub subtracts amounts of two decimal coins with same denom.
func (coin DecCoin) Sub(coinB DecCoin) DecCoin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("coin denom different: %v %v\n", coin.Denom, coinB.Denom))
	}
	res := DecCoin{coin.Denom, coin.Amount.Sub(coinB.Amount)}
	if res.IsNegative() {
		panic("negative decimal coin amount")
	}
	return res
}

// TruncateDecimal returns a Coin with a truncated decimal and a DecCoin for the
// change. Note, the change may be zero.
func (coin DecCoin) TruncateDecimal() (Coin, DecCoin) {
	truncated := coin.Amount.TruncateInt()
	change := coin.Amount.Sub(truncated.ToDec())
	return NewCoin(coin.Denom, truncated), NewDecCoinFromDec(coin.Denom, change)
}

// IsPositive returns true if coin amount is positive.
//
// TODO: Remove once unsigned integers are used.
func (coin DecCoin) IsPositive() bool {
	return coin.Amount.IsPositive()
}

// IsNegative returns true if the coin amount is negative and false otherwise.
//
// TODO: Remove once unsigned integers are used.
func (coin DecCoin) IsNegative() bool {
	return coin.Amount.Sign() == -1
}

// String implements the Stringer interface for DecCoin. It returns a
// human-readable representation of a decimal coin.
func (coin DecCoin) String() string {
	return fmt.Sprintf("%v%v", coin.Amount, coin.Denom)
}

// IsValid returns true if the DecCoin has a non-negative amount and the denom
// is valid.
func (coin DecCoin) IsValid() bool {
	if err := validateDenom(coin.Denom); err != nil {
		return false
	}
	return !coin.IsNegative()
}

// ----------------------------------------------------------------------------
// Decimal Coins

// DecCoins defines a slice of coins with decimal values
type DecCoins []DecCoin

// NewDecCoins constructs a new coin set with decimal values
// from regular Coins.
func NewDecCoins(coins Coins) DecCoins {
	decCoins := make(DecCoins, len(coins))
	newCoins := NewCoins(coins...)
	for i, coin := range newCoins {
		decCoins[i] = NewDecCoinFromCoin(coin)
	}

	return decCoins
}

// String implements the Stringer interface for DecCoins. It returns a
// human-readable representation of decimal coins.
func (coins DecCoins) String() string {
	if len(coins) == 0 {
		return ""
	}

	out := ""
	for _, coin := range coins {
		out += fmt.Sprintf("%v,", coin.String())
	}

	return out[:len(out)-1]
}

// TruncateDecimal returns the coins with truncated decimals and returns the
// change. Note, it will not return any zero-amount coins in either the truncated or
// change coins.
func (coins DecCoins) TruncateDecimal() (truncatedCoins Coins, changeCoins DecCoins) {
	for _, coin := range coins {
		truncated, change := coin.TruncateDecimal()
		if !truncated.IsZero() {
			truncatedCoins = truncatedCoins.Add(NewCoins(truncated))
		}
		if !change.IsZero() {
			changeCoins = changeCoins.Add(DecCoins{change})
		}
	}

	return truncatedCoins, changeCoins
}

// Add adds two sets of DecCoins.
//
// NOTE: Add operates under the invariant that coins are sorted by
// denominations.
//
// CONTRACT: Add will never return Coins where one Coin has a non-positive
// amount. In otherwords, IsValid will always return true.
func (coins DecCoins) Add(coinsB DecCoins) DecCoins {
	return coins.safeAdd(coinsB)
}

// safeAdd will perform addition of two DecCoins sets. If both coin sets are
// empty, then an empty set is returned. If only a single set is empty, the
// other set is returned. Otherwise, the coins are compared in order of their
// denomination and addition only occurs when the denominations match, otherwise
// the coin is simply added to the sum assuming it's not zero.
func (coins DecCoins) safeAdd(coinsB DecCoins) DecCoins {
	sum := ([]DecCoin)(nil)
	indexA, indexB := 0, 0
	lenA, lenB := len(coins), len(coinsB)

	for {
		if indexA == lenA {
			if indexB == lenB {
				// return nil coins if both sets are empty
				return sum
			}

			// return set B (excluding zero coins) if set A is empty
			return append(sum, removeZeroDecCoins(coinsB[indexB:])...)
		} else if indexB == lenB {
			// return set A (excluding zero coins) if set B is empty
			return append(sum, removeZeroDecCoins(coins[indexA:])...)
		}

		coinA, coinB := coins[indexA], coinsB[indexB]

		switch strings.Compare(coinA.Denom, coinB.Denom) {
		case -1: // coin A denom < coin B denom
			if !coinA.IsZero() {
				sum = append(sum, coinA)
			}

			indexA++

		case 0: // coin A denom == coin B denom
			res := coinA.Add(coinB)
			if !res.IsZero() {
				sum = append(sum, res)
			}

			indexA++
			indexB++

		case 1: // coin A denom > coin B denom
			if !coinB.IsZero() {
				sum = append(sum, coinB)
			}

			indexB++
		}
	}
}

// negative returns a set of coins with all amount negative.
func (coins DecCoins) negative() DecCoins {
	res := make([]DecCoin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Neg(),
		})
	}
	return res
}

// Sub subtracts a set of DecCoins from another (adds the inverse).
func (coins DecCoins) Sub(coinsB DecCoins) DecCoins {
	diff, hasNeg := coins.SafeSub(coinsB)
	if hasNeg {
		panic("negative coin amount")
	}

	return diff
}

// SafeSub performs the same arithmetic as Sub but returns a boolean if any
// negative coin amount was returned.
func (coins DecCoins) SafeSub(coinsB DecCoins) (DecCoins, bool) {
	diff := coins.safeAdd(coinsB.negative())
	return diff, diff.IsAnyNegative()
}

// Intersect will return a new set of coins which contains the minimum DecCoin
// for common denoms found in both `coins` and `coinsB`. For denoms not common
// to both `coins` and `coinsB` the minimum is considered to be 0, thus they
// are not added to the final set.In other words, trim any denom amount from
// coin which exceeds that of coinB, such that (coin.Intersect(coinB)).IsLTE(coinB).
func (coins DecCoins) Intersect(coinsB DecCoins) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		minCoin := DecCoin{
			Denom:  coin.Denom,
			Amount: MinDec(coin.Amount, coinsB.AmountOf(coin.Denom)),
		}
		res[i] = minCoin
	}
	return removeZeroDecCoins(res)
}

// IsAnyNegative returns true if there is at least one coin whose amount
// is negative; returns false otherwise. It returns false if the DecCoins set
// is empty too.
//
// TODO: Remove once unsigned integers are used.
func (coins DecCoins) IsAnyNegative() bool {
	for _, coin := range coins {
		if coin.IsNegative() {
			return true
		}
	}

	return false
}

// MulDec multiplies all the coins by a decimal.
//
// CONTRACT: No zero coins will be returned.
func (coins DecCoins) MulDec(d Dec) DecCoins {
	var res DecCoins
	for _, coin := range coins {
		product := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Mul(d),
		}

		if !product.IsZero() {
			res = res.Add(DecCoins{product})
		}
	}

	return res
}

// MulDecTruncate multiplies all the decimal coins by a decimal, truncating. It
// panics if d is zero.
//
// CONTRACT: No zero coins will be returned.
func (coins DecCoins) MulDecTruncate(d Dec) DecCoins {
	var res DecCoins

	for _, coin := range coins {
		product := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.MulTruncate(d),
		}

		if !product.IsZero() {
			res = res.Add(DecCoins{product})
		}
	}

	return res
}

// QuoDec divides all the decimal coins by a decimal. It panics if d is zero.
//
// CONTRACT: No zero coins will be returned.
func (coins DecCoins) QuoDec(d Dec) DecCoins {
	if d.IsZero() {
		panic("invalid zero decimal")
	}

	var res DecCoins
	for _, coin := range coins {
		quotient := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Quo(d),
		}

		if !quotient.IsZero() {
			res = res.Add(DecCoins{quotient})
		}
	}

	return res
}

// QuoDecTruncate divides all the decimal coins by a decimal, truncating. It
// panics if d is zero.
//
// CONTRACT: No zero coins will be returned.
func (coins DecCoins) QuoDecTruncate(d Dec) DecCoins {
	if d.IsZero() {
		panic("invalid zero decimal")
	}

	var res DecCoins
	for _, coin := range coins {
		quotient := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.QuoTruncate(d),
		}

		if !quotient.IsZero() {
			res = res.Add(DecCoins{quotient})
		}
	}

	return res
}

// Empty returns true if there are no coins and false otherwise.
func (coins DecCoins) Empty() bool {
	return len(coins) == 0
}

// AmountOf returns the amount of a denom from deccoins
func (coins DecCoins) AmountOf(denom string) Dec {
	mustValidateDenom(denom)

	switch len(coins) {
	case 0:
		return ZeroDec()

	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return coin.Amount
		}
		return ZeroDec()

	default:
		midIdx := len(coins) / 2 // 2:1, 3:1, 4:2
		coin := coins[midIdx]

		if denom < coin.Denom {
			return coins[:midIdx].AmountOf(denom)
		} else if denom == coin.Denom {
			return coin.Amount
		} else {
			return coins[midIdx+1:].AmountOf(denom)
		}
	}
}

// IsEqual returns true if the two sets of DecCoins have the same value.
func (coins DecCoins) IsEqual(coinsB DecCoins) bool {
	if len(coins) != len(coinsB) {
		return false
	}

	coins = coins.Sort()
	coinsB = coinsB.Sort()

	for i := 0; i < len(coins); i++ {
		if !coins[i].IsEqual(coinsB[i]) {
			return false
		}
	}

	return true
}

// IsZero returns whether all coins are zero
func (coins DecCoins) IsZero() bool {
	for _, coin := range coins {
		if !coin.Amount.IsZero() {
			return false
		}
	}
	return true
}

// IsValid asserts the DecCoins are sorted, have positive amount, and Denom
// does not contain upper case characters.
func (coins DecCoins) IsValid() bool {
	switch len(coins) {
	case 0:
		return true

	case 1:
		if err := validateDenom(coins[0].Denom); err != nil {
			return false
		}
		return coins[0].IsPositive()

	default:
		// check single coin case
		if !(DecCoins{coins[0]}).IsValid() {
			return false
		}

		lowDenom := coins[0].Denom
		for _, coin := range coins[1:] {
			if strings.ToLower(coin.Denom) != coin.Denom {
				return false
			}
			if coin.Denom <= lowDenom {
				return false
			}
			if !coin.IsPositive() {
				return false
			}

			// we compare each coin against the last denom
			lowDenom = coin.Denom
		}

		return true
	}
}

// IsAllPositive returns true if there is at least one coin and all currencies
// have a positive value.
//
// TODO: Remove once unsigned integers are used.
func (coins DecCoins) IsAllPositive() bool {
	if len(coins) == 0 {
		return false
	}

	for _, coin := range coins {
		if !coin.IsPositive() {
			return false
		}
	}

	return true
}

func removeZeroDecCoins(coins DecCoins) DecCoins {
	i, l := 0, len(coins)
	for i < l {
		if coins[i].IsZero() {
			// remove coin
			coins = append(coins[:i], coins[i+1:]...)
			l--
		} else {
			i++
		}
	}

	return coins[:i]
}

//-----------------------------------------------------------------------------
// Sorting

var _ sort.Interface = Coins{}

//nolint
func (coins DecCoins) Len() int           { return len(coins) }
func (coins DecCoins) Less(i, j int) bool { return coins[i].Denom < coins[j].Denom }
func (coins DecCoins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

// Sort is a helper function to sort the set of decimal coins in-place.
func (coins DecCoins) Sort() DecCoins {
	sort.Sort(coins)
	return coins
}

// ----------------------------------------------------------------------------
// Parsing

// ParseDecCoin parses a decimal coin from a string, returning an error if
// invalid. An empty string is considered invalid.
func ParseDecCoin(coinStr string) (coin DecCoin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reDecCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return DecCoin{}, fmt.Errorf("invalid decimal coin expression: %s", coinStr)
	}

	amountStr, denomStr := matches[1], matches[2]

	amount, err := NewDecFromStr(amountStr)
	if err != nil {
		return DecCoin{}, errors.Wrap(err, fmt.Sprintf("failed to parse decimal coin amount: %s", amountStr))
	}

	if err := validateDenom(denomStr); err != nil {
		return DecCoin{}, fmt.Errorf("invalid denom cannot contain upper case characters or spaces: %s", err)
	}

	return NewDecCoinFromDec(denomStr, amount), nil
}

// ParseDecCoins will parse out a list of decimal coins separated by commas.
// If nothing is provided, it returns nil DecCoins. Returned decimal coins are
// sorted.
func ParseDecCoins(coinsStr string) (DecCoins, error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	coinStrs := strings.Split(coinsStr, ",")
	coins := make(DecCoins, len(coinStrs))
	for i, coinStr := range coinStrs {
		coin, err := ParseDecCoin(coinStr)
		if err != nil {
			return nil, err
		}

		coins[i] = coin
	}

	// sort coins for determinism
	coins.Sort()

	// validate coins before returning
	if !coins.IsValid() {
		return nil, fmt.Errorf("parsed decimal coins are invalid: %#v", coins)
	}

	return coins, nil
}
