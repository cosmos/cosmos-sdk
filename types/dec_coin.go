package types

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
// Decimal Coin

// Coins which can have additional decimal points
type DecCoin struct {
	Denom  string `json:"denom"`
	Amount Dec    `json:"amount"`
}

func NewDecCoin(denom string, amount Int) DecCoin {
	validateDenom(denom)

	if amount.LT(ZeroInt()) {
		panic(fmt.Sprintf("negative coin amount: %v\n", amount))
	}

	return DecCoin{
		Denom:  denom,
		Amount: NewDecFromInt(amount),
	}
}

func NewDecCoinFromDec(denom string, amount Dec) DecCoin {
	validateDenom(denom)

	if amount.LT(ZeroDec()) {
		panic(fmt.Sprintf("negative decimal coin amount: %v\n", amount))
	}

	return DecCoin{
		Denom:  denom,
		Amount: amount,
	}
}

func NewDecCoinFromCoin(coin Coin) DecCoin {
	if coin.Amount.LT(ZeroInt()) {
		panic(fmt.Sprintf("negative decimal coin amount: %v\n", coin.Amount))
	}
	if strings.ToLower(coin.Denom) != coin.Denom {
		panic(fmt.Sprintf("denom cannot contain upper case characters: %s\n", coin.Denom))
	}

	return DecCoin{
		Denom:  coin.Denom,
		Amount: NewDecFromInt(coin.Amount),
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

// Adds amounts of two coins with same denom
func (coin DecCoin) Plus(coinB DecCoin) DecCoin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("coin denom different: %v %v\n", coin.Denom, coinB.Denom))
	}
	return DecCoin{coin.Denom, coin.Amount.Add(coinB.Amount)}
}

// Subtracts amounts of two coins with same denom
func (coin DecCoin) Minus(coinB DecCoin) DecCoin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("coin denom different: %v %v\n", coin.Denom, coinB.Denom))
	}
	return DecCoin{coin.Denom, coin.Amount.Sub(coinB.Amount)}
}

// return the decimal coins with trunctated decimals, and return the change
func (coin DecCoin) TruncateDecimal() (Coin, DecCoin) {
	truncated := coin.Amount.TruncateInt()
	change := coin.Amount.Sub(NewDecFromInt(truncated))
	return NewCoin(coin.Denom, truncated), DecCoin{coin.Denom, change}
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

// ----------------------------------------------------------------------------
// Decimal Coins

// coins with decimal
type DecCoins []DecCoin

func NewDecCoins(coins Coins) DecCoins {
	dcs := make(DecCoins, len(coins))
	for i, coin := range coins {
		dcs[i] = NewDecCoinFromCoin(coin)
	}
	return dcs
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
// change.
func (coins DecCoins) TruncateDecimal() (Coins, DecCoins) {
	changeSum := DecCoins{}
	out := make(Coins, len(coins))

	for i, coin := range coins {
		truncated, change := coin.TruncateDecimal()
		out[i] = truncated
		changeSum = changeSum.Plus(DecCoins{change})
	}

	return out, changeSum
}

// Plus adds two sets of DecCoins.
//
// NOTE: Plus operates under the invariant that coins are sorted by
// denominations.
//
// CONTRACT: Plus will never return Coins where one Coin has a non-positive
// amount. In otherwords, IsValid will always return true.
func (coins DecCoins) Plus(coinsB DecCoins) DecCoins {
	return coins.safePlus(coinsB)
}

// safePlus will perform addition of two DecCoins sets. If both coin sets are
// empty, then an empty set is returned. If only a single set is empty, the
// other set is returned. Otherwise, the coins are compared in order of their
// denomination and addition only occurs when the denominations match, otherwise
// the coin is simply added to the sum assuming it's not zero.
func (coins DecCoins) safePlus(coinsB DecCoins) DecCoins {
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
			res := coinA.Plus(coinB)
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

// Minus subtracts a set of DecCoins from another (adds the inverse).
func (coins DecCoins) Minus(coinsB DecCoins) DecCoins {
	diff, hasNeg := coins.SafeMinus(coinsB)
	if hasNeg {
		panic("negative coin amount")
	}

	return diff
}

// SafeMinus performs the same arithmetic as Minus but returns a boolean if any
// negative coin amount was returned.
func (coins DecCoins) SafeMinus(coinsB DecCoins) (DecCoins, bool) {
	diff := coins.safePlus(coinsB.negative())
	return diff, diff.IsAnyNegative()
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

// multiply all the coins by a decimal
func (coins DecCoins) MulDec(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		product := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Mul(d),
		}
		res[i] = product
	}
	return res
}

// multiply all the coins by a decimal, truncating
func (coins DecCoins) MulDecTruncate(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		product := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.MulTruncate(d),
		}
		res[i] = product
	}
	return res
}

// divide all the coins by a decimal
func (coins DecCoins) QuoDec(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		quotient := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Quo(d),
		}
		res[i] = quotient
	}
	return res
}

// divide all the coins by a decimal, truncating
func (coins DecCoins) QuoDecTruncate(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		quotient := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.QuoTruncate(d),
		}
		res[i] = quotient
	}
	return res
}

// Empty returns true if there are no coins and false otherwise.
func (coins DecCoins) Empty() bool {
	return len(coins) == 0
}

// returns the amount of a denom from deccoins
func (coins DecCoins) AmountOf(denom string) Dec {
	validateDenom(denom)

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

// return whether all coins are zero
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
		if strings.ToLower(coins[0].Denom) != coins[0].Denom {
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

	if denomStr != strings.ToLower(denomStr) {
		return DecCoin{}, fmt.Errorf("denom cannot contain upper case characters: %s", denomStr)
	}

	return NewDecCoinFromDec(denomStr, amount), nil
}

// ParseDecCoins will parse out a list of decimal coins separated by commas.
// If nothing is provided, it returns nil DecCoins. Returned decimal coins are
// sorted.
func ParseDecCoins(coinsStr string) (coins DecCoins, err error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	splitRe := regexp.MustCompile(",|;")
	coinStrs := splitRe.Split(coinsStr, -1)
	for _, coinStr := range coinStrs {
		coin, err := ParseDecCoin(coinStr)
		if err != nil {
			return nil, err
		}

		coins = append(coins, coin)
	}

	// sort coins for determinism
	coins.Sort()

	// validate coins before returning
	if !coins.IsValid() {
		return nil, fmt.Errorf("parsed decimal coins are invalid: %#v", coins)
	}

	return coins, nil
}
