package types

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

//-----------------------------------------------------------------------------
// Coin

// Coin hold some amount of one currency.
//
// CONTRACT: A coin will never hold a negative amount of any denomination.
//
// TODO: Make field members private for further safety.
type Coin struct {
	Denom  string `json:"denom"`
	Amount Uint   `json:"amount"`
}

// NewUint64Coin returns a new coin with a denomination and amount.
func NewUint64Coin(denom string, amount uint64) Coin {
	return NewCoin(denom, NewUint(amount))
}

// NewCoin returns a new coin with a denomination and amount. It will panic if
// the amount is negative.
func NewCoin(denom string, amount Uint) Coin {
	validateDenom(denom)

	// This should never happen
	if amount.LT(ZeroUint()) {
		panic(fmt.Sprintf("negative coin amount: %v\n", amount))
	}

	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

// String provides a human-readable representation of a coin
func (coin Coin) String() string { return fmt.Sprintf("%v%v", coin.Amount, coin.Denom) }

// IsZero returns if this represents no money
func (coin Coin) IsZero() bool { return coin.Amount.IsZero() }

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (coin Coin) IsGTE(other Coin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}

	return !coin.Amount.LT(other.Amount)
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (coin Coin) IsLT(other Coin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}

	return coin.Amount.LT(other.Amount)
}

// IsEqual returns true if the two sets of Coins have the same value
func (coin Coin) IsEqual(other Coin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}

	return coin.Amount.Equal(other.Amount)
}

// Adds amounts of two coins with same denom. If the coins differ in denom then
// it panics.
func (coin Coin) Plus(coinB Coin) Coin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}

	return NewCoin(coin.Denom, coin.Amount.Add(coinB.Amount))
}

// Subtracts amounts of two coins with same denom. If the coins differ in denom
// then it panics.
func (coin Coin) Minus(coinB Coin) Coin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}

	return NewCoin(coin.Denom, coin.Amount.Sub(coinB.Amount))
}

// IsPositive returns true if coin amount is positive.
//
// TODO: Remove once unsigned integers are used.
func (coin Coin) IsPositive() bool {
	return !coin.Amount.IsZero()
}

//-----------------------------------------------------------------------------
// Coins

// Coins is a set of Coin, one per currency
type Coins []Coin

func NewCoins(coins ...Coin) Coins {
	// Remove zeroes
	newCoins := removeZeroCoins(Coins(coins))
	if len(newCoins) == 0 {
		return Coins{}
	}

	// Sort
	newCoins.Sort()

	// Detect duplicate Denoms
	if dupIndex := findDup(newCoins); dupIndex != -1 {
		panic(fmt.Errorf("find duplicate denom: %s", newCoins[dupIndex]))
	}

	// Validate
	if !newCoins.IsValid() {
		panic(fmt.Errorf("invalid coin set: %s", newCoins))
	}

	return newCoins
}

func NewCoinsFromDenomAmountPairs(denoms []string, amounts []Uint) Coins {
	if len(denoms) != len(amounts) {
		panic("equal number of denoms and amounts must be supplied")
	}
	coins := make([]Coin, len(denoms))
	for i := 0; i < len(denoms); i++ {
		coins[i] = NewCoin(denoms[i], amounts[i])
	}
	return NewCoins(coins...)
}

// this work on the assumption that coins are sorted
func findDup(coins Coins) int {
	if len(coins) <= 1 {
		return -1
	}

	prevDenom := coins[0]
	for i := 1; i < len(coins); i++ {
		if coins[i] == prevDenom {
			return i
		}
	}

	return -1
}

func (coins Coins) String() string {
	if coins.IsZero() {
		return ""
	}
	coins.Sort()

	out := make([]string, len(coins))
	for i, coin := range coins {
		out[i] = coin.String()
	}
	return strings.Join(out, ",")
}

// IsValid asserts the Coins are sorted, have positive amount,
// and Denom does not contain upper case characters.
func (coins Coins) IsValid() bool {
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
		if !(Coins{coins[0]}).IsValid() {
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

// Plus adds two sets of coins.
//
// e.g.
// {2A} + {A, 2B} = {3A, 2B}
// {2A} + {0B} = {2A}
//
// NOTE: Plus operates under the invariant that coins are sorted by
// denominations.
//
// CONTRACT: Plus will never return Coins where one Coin has a non-positive
// amount. In otherwords, IsValid will always return true.
func (coins Coins) Plus(coinsB Coins) Coins { return coins.unsafeSum(coinsB, false) }

// Minus subtracts a set of coins from another.
//
// e.g.
// {2A, 3B} - {A} = {A, 3B}
// {2A} - {0B} = {2A}
// {A, B} - {A} = {B}
//
// CONTRACT: Minus will never return Coins where one Coin has a non-positive
// amount. In otherwords, IsValid will always return true.
func (coins Coins) Minus(coinsB Coins) Coins {
	res, err := coins.SafeMinus(coinsB)
	if err != nil {
		panic(err)
	}
	return res
}

// SafeMinus carries out preemptive checks for negative results before attempting
// to subtract coinsB from coinsA: it checks whether coinsB's denoms are included
// in coinsA and their amount are less than or equal to their respective denoms
// in coinsA. Both conditions would cause Minus to panic. Returns coinA - coinsB
// result set and true if it the aforementioned conditions are met; returns nil
// Coins and false otherwise.
func (coins Coins) SafeMinus(coinsB Coins) (Coins, error) {
	if !coinsB.IsAllLTE(coins) {
		return ZeroCoins(), errors.New("result contains negative amounts")
	}
	return coins.unsafeSum(coinsB, true), nil
}

// ZeroCoins returns an empty Coins.
func ZeroCoins() Coins { return ([]Coin)(nil) }

// safePlus will perform addition of two coins sets. If both coin sets are
// empty, then an empty set is returned. If only a single set is empty, the
// other set is returned. Otherwise, the coins are compared in order of their
// denomination and addition only occurs when the denominations match, otherwise
// the coin is simply added to the sum assuming it's not zero.
func (coins Coins) unsafeSum(coinsB Coins, subOp bool) Coins {
	// no-ops
	if coinsB.IsZero() {
		return copyCoins(coins)
	}
	if coins.IsZero() && !subOp {
		return copyCoins(coinsB)
	}

	sum := ZeroCoins()
	indexA, indexB := 0, 0
	lenA, lenB := len(coins), len(coinsB)

	for {
		if indexA == lenA {
			if indexB == lenB {
				// return nil coins if both sets are empty
				return sum
			}

			// return set B (excluding zero coins) if set A is empty
			return append(sum, removeZeroCoins(coinsB[indexB:])...)
		} else if indexB == lenB {
			// return set A (excluding zero coins) if set B is empty
			return append(sum, removeZeroCoins(coins[indexA:])...)
		}

		coinA, coinB := coins[indexA], coinsB[indexB]

		switch strings.Compare(coinA.Denom, coinB.Denom) {
		case -1: // coin A denom < coin B denom
			if !coinA.IsZero() {
				sum = append(sum, coinA)
			}

			indexA++

		case 0: // coin A denom == coin B denom
			var res Coin
			if subOp {
				res = coinA.Minus(coinB)
			} else {
				res = coinA.Plus(coinB)
			}
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

// ContainsDenomsOf returns true if coinsB' denom set
// is subset of the receiver's denoms.
func (coins Coins) ContainsDenomsOf(coinsB Coins) bool {
	// more denoms in B than in receiver
	if len(coinsB) > len(coins) {
		return false
	}

	for _, coinB := range coinsB {
		if coins.AmountOf(coinB.Denom).IsZero() {
			return false
		}
	}

	return true
}

// SameDenomsOf returns true if both Coin sets have the
// very same denoms.
func (coins Coins) SameDenomsOf(coinsB Coins) bool {
	return (len(coins) == len(coinsB)) && coins.ContainsDenomsOf(coinsB)
}

// IsAllGT returns true if for every denom in coins, the denom is present at a
// greater amount in coinsB.
func (coins Coins) IsAllGT(coinsB Coins) bool {
	if len(coins) == 0 {
		return false
	}

	if len(coinsB) == 0 {
		return true
	}

	if !coins.ContainsDenomsOf(coinsB) {
		return false
	}

	for _, coinB := range coinsB {
		amountA, amountB := coins.AmountOf(coinB.Denom), coinB.Amount
		if !amountA.GT(amountB) {
			return false
		}
	}

	return true
}

// IsAllGT returns false if for any denom in coins,
// the denom is present in coinsB at a greater or
// equal amount
func (coins Coins) IsAllGTE(coinsB Coins) bool {
	if len(coinsB) == 0 {
		return true
	}

	if len(coins) == 0 {
		return false
	}

	if !coinsB.Difference(coins).IsZero() {
		return false
	}

	for _, coin := range coins {
		if coin.Amount.LT(coinsB.AmountOf(coin.Denom)) {
			return false
		}
	}

	return true
}

// IsAllLT returns true if all denoms in coins
// are less than their counterparts in coinsB.
func (coins Coins) IsAllLT(coinsB Coins) bool {
	// coinsB is zero, nothing can be less than that
	if coinsB.IsZero() {
		return false
	}
	// if only coins is zero, then always return true
	if coins.IsZero() {
		return true
	}
	// alternatively compare the two sets
	for _, coin := range coins {
		if !coin.Amount.LT(coinsB.AmountOf(coin.Denom)) {
			return false
		}
	}
	return true
}

// IsAllLTE returns false iff if there's a denom in coinsB whose amount
// is greater than coins respective denom; returns true otherwise.
func (coins Coins) IsAllLTE(coinsB Coins) bool {
	// coins is an empty set, therefore
	// all items in coins are smaller than those in coinsB
	if coins.IsZero() {
		return true
	}
	// coinsB is empty set (coins is not empty), therefore
	// all items in coins are greater than those in coinsB
	if coinsB.IsZero() {
		return false
	}
	// if some denoms are in coins only, then return false
	if !coins.Difference(coinsB).IsZero() {
		return false
	}
	// compare the two sets
	for _, coin := range coins {
		if coin.Amount.GT(coinsB.AmountOf(coin.Denom)) {
			return false
		}
	}
	return true
}

// IsAnyGTE returns true iff coins contains at least one denom that is present
// at a greater or equal amount in coinsB; it returns false otherwise.
//
// NOTE: IsAnyGTE operates under the invariant that both coin sets are sorted
// by denominations and there exists no zero coins.
func (coins Coins) IsAnyGTE(coinsB Coins) bool {
	if coins.IsZero() {
		return false
	}

	if coinsB.IsZero() {
		return true
	}

	if !coins.Difference(coinsB).IsZero() {
		return true
	}

	for _, coin := range coins {
		if coin.Amount.GTE(coinsB.AmountOf(coin.Denom)) {
			return true
		}
	}

	return false
}

// IsZero returns true if there are no coins or all coins are zero.
func (coins Coins) IsZero() bool {
	if len(coins) == 0 {
		return true
	}
	for _, coin := range coins {
		if coin.IsZero() {
			panic("sanity check failed: Coins cannot contain zero-Coin")
		}
	}
	return false
}

// IsEqual returns true if the two sets of Coins have the same value
func (coins Coins) IsEqual(coinsB Coins) bool {
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

// Empty returns true if there are no coins and false otherwise.
func (coins Coins) Empty() bool {
	return len(coins) == 0
}

// Returns the amount of a denom from coins
func (coins Coins) AmountOf(denom string) Uint {
	validateDenom(denom)

	switch len(coins) {
	case 0:
		return ZeroUint()

	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return coin.Amount
		}
		return ZeroUint()

	default:
		midIdx := len(coins) / 2 // 2:1, 3:1, 4:2
		coin := coins[midIdx]

		if denom == coin.Denom {
			return coin.Amount
		}
		if denom < coin.Denom {
			return coins[:midIdx].AmountOf(denom)
		}
		return coins[midIdx+1:].AmountOf(denom)
	}
}

// IsAllPositive returns true if there is at least one coin and all currencies
// have a positive value.
//
// TODO: Remove once unsigned integers are used.
func (coins Coins) IsAllPositive() bool {
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

//-----------------------------------------------------------------------------
// Sort interface

//nolint
func (coins Coins) Len() int           { return len(coins) }
func (coins Coins) Less(i, j int) bool { return coins[i].Denom < coins[j].Denom }
func (coins Coins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

var _ sort.Interface = Coins{}

// Sort is a helper function to sort the set of coins inplace
func (coins Coins) Sort() Coins {
	sort.Sort(coins)
	return coins
}

// Difference returns all coins in A whose denom is
// present at a nonzero value in coinsB.
func (coins Coins) Difference(coinsB Coins) Coins {
	if len(coins) == 0 {
		return ZeroCoins()
	}

	if len(coinsB) == 0 {
		return copyCoins(coins)
	}

	res := ZeroCoins()
	for _, coin := range coins {
		if coinsB.AmountOf(coin.Denom).IsZero() {
			res = append(res, coin)
		}
	}

	return res
}

//-----------------------------------------------------------------------------
// Parsing

var (
	// Denominations can be 3 ~ 16 characters long.
	reDnm     = `[[:alpha:]][[:alnum:]]{2,15}`
	reAmt     = `[[:digit:]]+`
	reDecAmt  = `[[:digit:]]*\.[[:digit:]]+`
	reSpc     = `[[:space:]]*`
	reCoin    = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmt, reSpc, reDnm))
	reDecCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmt, reSpc, reDnm))
)

// ParseCoin parses a cli input for one coin type, returning errors if invalid.
// This returns an error on an empty string as well.
func ParseCoin(coinStr string) (coin Coin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return Coin{}, fmt.Errorf("invalid coin expression: %s", coinStr)
	}

	denomStr, amountStr := matches[2], matches[1]

	amount, err := ParseUint(amountStr)
	if err != nil {
		return Coin{}, fmt.Errorf("failed to parse coin amount: %s", err)
	}

	if denomStr != strings.ToLower(denomStr) {
		return Coin{}, fmt.Errorf("denom cannot contain upper case characters: %s", denomStr)
	}

	return NewCoin(denomStr, amount), nil
}

// ParseCoins will parse out a list of coins separated by commas.
// If nothing is provided, it returns nil Coins.
// Returned coins are sorted.
func ParseCoins(coinsStr string) (coins Coins, err error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	coinStrs := strings.Split(coinsStr, ",")
	for _, coinStr := range coinStrs {
		coin, err := ParseCoin(coinStr)
		if err != nil {
			return nil, err
		}
		coins = append(coins, coin)
	}

	// Sort coins for determinism.
	coins.Sort()

	// Validate coins before returning.
	if !coins.IsValid() {
		return nil, fmt.Errorf("parseCoins invalid: %#v", coins)
	}

	return coins, nil
}

// MustParseCoins panics if the underlying ParseCoins call fails.
func MustParseCoins(s string) Coins {
	c, err := ParseCoins(s)
	if err != nil {
		panic(err)
	}
	return c
}

func validateDenom(denom string) {
	if len(denom) < 3 || len(denom) > 16 {
		panic(fmt.Sprintf("invalid denom length: %s", denom))
	}
	if strings.ToLower(denom) != denom {
		panic(fmt.Sprintf("denom cannot contain upper case characters: %s", denom))
	}
}

// removeZeroCoins removes all zero coins from the given coin set in-place.
func removeZeroCoins(coins Coins) Coins {
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

func copyCoins(coins Coins) Coins {
	copyCoins := make(Coins, len(coins))
	copy(copyCoins, coins)
	return copyCoins
}
