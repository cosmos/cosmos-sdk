package types

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var regexAnySpace = regexp.MustCompile(`\s`)

//-----------------------------------------------------------------------------
// Coin

// Coin hold some amount of one currency.
//
// CONTRACT: A coin will never hold a negative amount of any denomination.
//
// TODO: Make field members private for further safety.
type Coin struct {
	Denom string `json:"denom"`

	// To allow the use of unsigned integers (see: #1273) a larger refactor will
	// need to be made. So we use signed integers for now with safety measures in
	// place preventing negative values being used.
	Amount Int `json:"amount"`
}

// NewCoin returns a new coin with a denomination and amount.
// It will panic if the amount is negative.
func NewCoin(denom string, amount Int) Coin {
	c := Coin{Denom: denom, Amount: amount}
	if err := c.Validate(false); err != nil {
		panic(err)
	}

	return c
}

// NewCoin returns a new coin with a denomination and amount.
// It will panic if the amount is less than or equal to zero.
func NewPositiveCoin(denom string, amount Int) Coin {
	c := Coin{Denom: denom, Amount: amount}
	if err := c.Validate(true); err != nil {
		panic(err)
	}

	return c
}

// NewInt64Coin returns a new coin with a denomination and amount.
// It will panic if the amount is negative.
func NewInt64Coin(denom string, amount int64) Coin {
	return NewCoin(denom, NewInt(amount))
}

// NewInt64Coin returns a new coin with a denomination and amount.
// It will panic if the amount less than or equal to zero.
func NewPositiveInt64Coin(denom string, amount int64) Coin {
	return NewPositiveCoin(denom, NewInt(amount))
}

// Validate validates coin's Amount and Denom. If strict is true, then
// it returns an error if Amount less than or equal to zero.
// If strict is false, then it returns an error if and only if Amount
// is less than zero.
func (coin Coin) Validate(strict bool) error {
	if err := validateIntCoinAmount(coin.Amount, strict); err != nil {
		return fmt.Errorf("%s: %s", err, coin.Amount)
	}

	if err := validateCoinDenomContainsSpace(coin.Denom); err != nil {
		return fmt.Errorf("%s: %s", err, coin.Denom)
	}

	if err := validateCoinDenomCase(coin.Denom); err != nil {
		return fmt.Errorf("%s: %s", err, coin.Denom)
	}

	return nil
}

// String provides a human-readable representation of a coin
func (coin Coin) String() string {
	return fmt.Sprintf("%v%v", coin.Amount, coin.Denom)
}

// SameDenomAs returns true if the two coins are the same denom
func (coin Coin) SameDenomAs(other Coin) bool {
	return (coin.Denom == other.Denom)
}

// IsZero returns if this represents no money
func (coin Coin) IsZero() bool {
	return coin.Amount.IsZero()
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (coin Coin) IsGTE(other Coin) bool {
	return coin.SameDenomAs(other) && (!coin.Amount.LT(other.Amount))
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (coin Coin) IsLT(other Coin) bool {
	return coin.SameDenomAs(other) && coin.Amount.LT(other.Amount)
}

// IsEqual returns true if the two sets of Coins have the same value
func (coin Coin) IsEqual(other Coin) bool {
	return coin.SameDenomAs(other) && (coin.Amount.Equal(other.Amount))
}

// Adds amounts of two coins with same denom. If the coins differ in denom then
// it panics.
func (coin Coin) Plus(coinB Coin) Coin {
	if !coin.SameDenomAs(coinB) {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}

	return Coin{coin.Denom, coin.Amount.Add(coinB.Amount)}
}

// Subtracts amounts of two coins with same denom. If the coins differ in denom
// then it panics.
func (coin Coin) Minus(coinB Coin) Coin {
	if !coin.SameDenomAs(coinB) {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}

	res := Coin{coin.Denom, coin.Amount.Sub(coinB.Amount)}
	if res.IsNegative() {
		panic("negative count amount")
	}

	return res
}

// IsPositive returns true if coin amount is positive.
//
// TODO: Remove once unsigned integers are used.
func (coin Coin) IsPositive() bool {
	return coin.Amount.Sign() == 1
}

// IsNegative returns true if the coin amount is negative and false otherwise.
//
// TODO: Remove once unsigned integers are used.
func (coin Coin) IsNegative() bool {
	return coin.Amount.Sign() == -1
}

//-----------------------------------------------------------------------------
// Coins

// Coins is a set of Coin, one per currency
type Coins []Coin

func (coins Coins) String() string {
	if len(coins) == 0 {
		return ""
	}

	out := ""
	for _, coin := range coins {
		out += fmt.Sprintf("%v,", coin.String())
	}
	return out[:len(out)-1]
}

// IsValid asserts the Coins are sorted, have positive amount,
// and Denom does not contain upper case characters.
func (coins Coins) IsValid() bool {
	switch len(coins) {
	case 0:
		return true
	case 1:
		if err := coins[0].Validate(true); err != nil {
			return false
		}
		return true
	default:
		// check single coin case
		if !(Coins{coins[0]}).IsValid() {
			return false
		}

		lowDenom := coins[0].Denom
		for _, coin := range coins[1:] {
			if err := coin.Validate(true); err != nil {
				return false
			}
			if coin.Denom <= lowDenom {
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
func (coins Coins) Plus(coinsB Coins) Coins {
	return coins.safePlus(coinsB)
}

// safePlus will perform addition of two coins sets. If both coin sets are
// empty, then an empty set is returned. If only a single set is empty, the
// other set is returned. Otherwise, the coins are compared in order of their
// denomination and addition only occurs when the denominations match, otherwise
// the coin is simply added to the sum assuming it's not zero.
func (coins Coins) safePlus(coinsB Coins) Coins {
	sum := ([]Coin)(nil)
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
	diff, hasNeg := coins.SafeMinus(coinsB)
	if hasNeg {
		panic("negative coin amount")
	}

	return diff
}

// SafeMinus performs the same arithmetic as Minus but returns a boolean if any
// negative coin amount was returned.
func (coins Coins) SafeMinus(coinsB Coins) (Coins, bool) {
	diff := coins.safePlus(coinsB.negative())
	return diff, diff.IsAnyNegative()
}

// IsAllGT returns true if for every denom in coins, the denom is present at a
// greater amount in coinsB.
func (coins Coins) IsAllGT(coinsB Coins) bool {
	diff, _ := coins.SafeMinus(coinsB)
	if len(diff) == 0 {
		return false
	}

	return diff.IsPositive()
}

// IsAllGTE returns true iff for every denom in coins, the denom is present at
// an equal or greater amount in coinsB.
func (coins Coins) IsAllGTE(coinsB Coins) bool {
	diff, _ := coins.SafeMinus(coinsB)
	if len(diff) == 0 {
		return true
	}

	return !diff.IsAnyNegative()
}

// IsAllLT returns True iff for every denom in coins, the denom is present at
// a smaller amount in coinsB.
func (coins Coins) IsAllLT(coinsB Coins) bool {
	return coinsB.IsAllGT(coins)
}

// IsAllLTE returns true iff for every denom in coins, the denom is present at
// a smaller or equal amount in coinsB.
func (coins Coins) IsAllLTE(coinsB Coins) bool {
	return coinsB.IsAllGTE(coins)
}

// IsAnyGTE returns true iff coins contains at least one denom that is present
// at a greater or equal amount in coinsB; it returns false otherwise.
//
// NOTE: IsAnyGTE operates under the invariant that both coin sets are sorted
// by denominations and there exists no zero coins.
func (coins Coins) IsAnyGTE(coinsB Coins) bool {
	if len(coinsB) == 0 {
		return false
	}

	for _, coin := range coins {
		amt := coinsB.AmountOf(coin.Denom)
		if coin.Amount.GTE(amt) && !amt.IsZero() {
			return true
		}
	}

	return false
}

// IsZero returns true if there are no coins or all coins are zero.
func (coins Coins) IsZero() bool {
	for _, coin := range coins {
		if !coin.IsZero() {
			return false
		}
	}
	return true
}

// IsEqual returns true if the two sets of Coins have the same value
func (coins Coins) IsEqual(coinsB Coins) bool {
	if len(coins) != len(coinsB) {
		return false
	}

	coins = coins.Sort()
	coinsB = coinsB.Sort()

	for i := 0; i < len(coins); i++ {
		if coins[i].Denom != coinsB[i].Denom || !coins[i].Amount.Equal(coinsB[i].Amount) {
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
func (coins Coins) AmountOf(denom string) Int {
	if strings.ToLower(denom) != denom {
		panic(fmt.Sprintf("denom cannot contain upper case characters: %s\n", denom))
	}
	switch len(coins) {
	case 0:
		return ZeroInt()

	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return coin.Amount
		}
		return ZeroInt()

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

// IsPositive returns true if there is at least one coin and all currencies
// have a positive value.
//
// TODO: Remove once unsigned integers are used.
func (coins Coins) IsPositive() bool {
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

// IsAnyNegative returns true if there is at least one coin whose amount
// is negative; returns false otherwise. It returns false if the coin set
// is empty too.
// TODO: Remove once unsigned integers are used.
func (coins Coins) IsAnyNegative() bool {
	if len(coins) == 0 {
		return false
	}

	for _, coin := range coins {
		if coin.IsNegative() {
			return true
		}
	}

	return false
}

// negative returns a set of coins with all amount negative.
//
// TODO: Remove once unsigned integers are used.
func (coins Coins) negative() Coins {
	res := make([]Coin, 0, len(coins))

	for _, coin := range coins {
		res = append(res, Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Neg(),
		})
	}

	return res
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
func ParseCoin(coinStr string) (Coin, error) {
	coin, err := parseCoinString(coinStr)
	if err != nil {
		return Coin{}, fmt.Errorf("failed to parse coin: %s", err)
	}

	if err := coin.Validate(false); err != nil {
		return Coin{}, fmt.Errorf("validation error: %s", err)
	}

	return coin, nil
}

// ParseCoin parses a cli input for one coin type, returning errors if invalid.
// This returns an error on an empty string as well.
func ParsePositiveCoin(coinStr string) (Coin, error) {
	coin, err := parseCoinString(coinStr)
	if err != nil {
		return Coin{}, fmt.Errorf("failed to parse coin: %s", err)
	}

	if err := coin.Validate(true); err != nil {
		return Coin{}, fmt.Errorf("validation error: %s", err)
	}

	return coin, nil
}

func parseCoinString(coinStr string) (Coin, error) {
	coinStr = strings.TrimSpace(coinStr)
	matches := reCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return Coin{}, fmt.Errorf("invalid coin expression %q", coinStr)
	}

	denomStr, amountStr := matches[2], matches[1]
	amount, ok := NewIntFromString(amountStr)
	if !ok {
		return Coin{}, fmt.Errorf("failed to parse coin amount %q", amountStr)
	}

	return Coin{Denom: denomStr, Amount: amount}, nil
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

func validateCoinDenomCase(denom string) error {
	if denom != strings.ToLower(denom) {
		return errors.New("denom cannot contain upper case characters")
	}
	return nil
}

func validateCoinDenomContainsSpace(denom string) error {
	if regexAnySpace.MatchString(denom) {
		return errors.New("denom cannot contain space characters")
	}
	return nil
}

func validateIntCoinAmount(amount Int, strict bool) error {
	if strict && amount.LTE(ZeroInt()) {
		return errors.New("non-positive coin amount")
	}
	if !strict && amount.LT(ZeroInt()) {
		return errors.New("negative coin amount")
	}
	return nil
}
