package types

import (
	"encoding/json"
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
	Denom string `json:"denom"`

	// To allow the use of unsigned integers (see: #1273) a larger refactor will
	// need to be made. So we use signed integers for now with safety measures in
	// place preventing negative values being used.
	Amount Int `json:"amount"`
}

// NewCoin returns a new coin with a denomination and amount. It will panic if
// the amount is negative.
func NewCoin(denom string, amount Int) Coin {
	if err := validate(denom, amount); err != nil {
		panic(err)
	}

	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

// NewInt64Coin returns a new coin with a denomination and amount. It will panic
// if the amount is negative.
func NewInt64Coin(denom string, amount int64) Coin {
	return NewCoin(denom, NewInt(amount))
}

// String provides a human-readable representation of a coin
func (coin Coin) String() string {
	return fmt.Sprintf("%v%v", coin.Amount, coin.Denom)
}

// validate returns an error if the Coin has a negative amount or if
// the denom is invalid.
func validate(denom string, amount Int) error {
	if err := ValidateDenom(denom); err != nil {
		return err
	}

	if amount.IsNegative() {
		return fmt.Errorf("negative coin amount: %v", amount)
	}

	return nil
}

// IsValid returns true if the Coin has a non-negative amount and the denom is vaild.
func (coin Coin) IsValid() bool {
	if err := validate(coin.Denom, coin.Amount); err != nil {
		return false
	}
	return true
}

// IsZero returns if this represents no money
func (coin Coin) IsZero() bool {
	return coin.Amount.IsZero()
}

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
func (coin Coin) Add(coinB Coin) Coin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}

	return Coin{coin.Denom, coin.Amount.Add(coinB.Amount)}
}

// Subtracts amounts of two coins with same denom. If the coins differ in denom
// then it panics.
func (coin Coin) Sub(coinB Coin) Coin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}

	res := Coin{coin.Denom, coin.Amount.Sub(coinB.Amount)}
	if res.IsNegative() {
		panic("negative coin amount")
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

// NewCoins constructs a new coin set.
func NewCoins(coins ...Coin) Coins {
	// remove zeroes
	newCoins := removeZeroCoins(Coins(coins))
	if len(newCoins) == 0 {
		return Coins{}
	}

	newCoins.Sort()

	// detect duplicate Denoms
	if dupIndex := findDup(newCoins); dupIndex != -1 {
		panic(fmt.Errorf("find duplicate denom: %s", newCoins[dupIndex]))
	}

	if !newCoins.IsValid() {
		panic(fmt.Errorf("invalid coin set: %s", newCoins))
	}

	return newCoins
}

type coinsJSON Coins

// MarshalJSON implements a custom JSON marshaller for the Coins type to allow
// nil Coins to be encoded as an empty array.
func (coins Coins) MarshalJSON() ([]byte, error) {
	if coins == nil {
		return json.Marshal(coinsJSON(Coins{}))
	}

	return json.Marshal(coinsJSON(coins))
}

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
		if err := ValidateDenom(coins[0].Denom); err != nil {
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

// Add adds two sets of coins.
//
// e.g.
// {2A} + {A, 2B} = {3A, 2B}
// {2A} + {0B} = {2A}
//
// NOTE: Add operates under the invariant that coins are sorted by
// denominations.
//
// CONTRACT: Add will never return Coins where one Coin has a non-positive
// amount. In otherwords, IsValid will always return true.
func (coins Coins) Add(coinsB ...Coin) Coins {
	return coins.safeAdd(coinsB)
}

// safeAdd will perform addition of two coins sets. If both coin sets are
// empty, then an empty set is returned. If only a single set is empty, the
// other set is returned. Otherwise, the coins are compared in order of their
// denomination and addition only occurs when the denominations match, otherwise
// the coin is simply added to the sum assuming it's not zero.
func (coins Coins) safeAdd(coinsB Coins) Coins {
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

// DenomsSubsetOf returns true if receiver's denom set
// is subset of coinsB's denoms.
func (coins Coins) DenomsSubsetOf(coinsB Coins) bool {
	// more denoms in B than in receiver
	if len(coins) > len(coinsB) {
		return false
	}

	for _, coin := range coins {
		if coinsB.AmountOf(coin.Denom).IsZero() {
			return false
		}
	}

	return true
}

// Sub subtracts a set of coins from another.
//
// e.g.
// {2A, 3B} - {A} = {A, 3B}
// {2A} - {0B} = {2A}
// {A, B} - {A} = {B}
//
// CONTRACT: Sub will never return Coins where one Coin has a non-positive
// amount. In otherwords, IsValid will always return true.
func (coins Coins) Sub(coinsB Coins) Coins {
	diff, hasNeg := coins.SafeSub(coinsB)
	if hasNeg {
		panic("negative coin amount")
	}

	return diff
}

// SafeSub performs the same arithmetic as Sub but returns a boolean if any
// negative coin amount was returned.
func (coins Coins) SafeSub(coinsB Coins) (Coins, bool) {
	diff := coins.safeAdd(coinsB.negative())
	return diff, diff.IsAnyNegative()
}

// IsAllGT returns true if for every denom in coinsB,
// the denom is present at a greater amount in coins.
func (coins Coins) IsAllGT(coinsB Coins) bool {
	if len(coins) == 0 {
		return false
	}

	if len(coinsB) == 0 {
		return true
	}

	if !coinsB.DenomsSubsetOf(coins) {
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

// IsAllGTE returns false if for any denom in coinsB,
// the denom is present at a smaller amount in coins;
// else returns true.
func (coins Coins) IsAllGTE(coinsB Coins) bool {
	if len(coinsB) == 0 {
		return true
	}

	if len(coins) == 0 {
		return false
	}

	for _, coinB := range coinsB {
		if coinB.Amount.GT(coins.AmountOf(coinB.Denom)) {
			return false
		}
	}

	return true
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

// IsAnyGT returns true iff for any denom in coins, the denom is present at a
// greater amount in coinsB.
//
// e.g.
// {2A, 3B}.IsAnyGT{A} = true
// {2A, 3B}.IsAnyGT{5C} = false
// {}.IsAnyGT{5C} = false
// {2A, 3B}.IsAnyGT{} = false
func (coins Coins) IsAnyGT(coinsB Coins) bool {
	if len(coinsB) == 0 {
		return false
	}

	for _, coin := range coins {
		amt := coinsB.AmountOf(coin.Denom)
		if coin.Amount.GT(amt) && !amt.IsZero() {
			return true
		}
	}

	return false
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
func (coins Coins) AmountOf(denom string) Int {
	mustValidateDenom(denom)

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
		switch {
		case denom < coin.Denom:
			return coins[:midIdx].AmountOf(denom)
		case denom == coin.Denom:
			return coin.Amount
		default:
			return coins[midIdx+1:].AmountOf(denom)
		}
	}
}

// GetDenomByIndex returns the Denom of the certain coin to make the findDup generic
func (coins Coins) GetDenomByIndex(i int) string {
	return coins[i].Denom
}

// IsAllPositive returns true if there is at least one coin and all currencies
// have a positive value.
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

// IsAnyNegative returns true if there is at least one coin whose amount
// is negative; returns false otherwise. It returns false if the coin set
// is empty too.
//
// TODO: Remove once unsigned integers are used.
func (coins Coins) IsAnyNegative() bool {
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
	reDnmString = `[a-z][a-z0-9]{2,15}|\x{0023}|[
\x{002A}|
\x{0030}-\x{0039}|
\x{00A9}|
\x{00AE}|
\x{203C}|
\x{2049}|
\x{2122}|
\x{2139}|
\x{2194}-\x{2199}|
\x{21A9}-\x{21AA}|
\x{231A}-\x{231B}|
\x{2328}|
\x{23CF}|
\x{23E9}-\x{23F3}|
\x{23F8}-\x{23FA}|
\x{24C2}|
\x{25AA}-\x{25AB}|
\x{25B6}|
\x{25C0}|
\x{25FB}-\x{25FE}|
\x{2600}-\x{2604}|
\x{260E}|
\x{2611}|
\x{2614}-\x{2615}|
\x{2618}|
\x{261D}|
\x{2620}|
\x{2622}-\x{2623}|
\x{2626}|
\x{262A}|
\x{262E}-\x{262F}|
\x{2638}-\x{263A}|
\x{2648}-\x{2653}|
\x{2660}|
\x{2663}|
\x{2665}-\x{2666}|
\x{2668}|
\x{267B}|
\x{267F}|
\x{2692}-\x{2694}|
\x{2696}-\x{2697}|
\x{2699}|
\x{269B}-\x{269C}|
\x{26A0}-\x{26A1}|
\x{26AA}-\x{26AB}|
\x{26B0}-\x{26B1}|
\x{26BD}-\x{26BE}|
\x{26C4}-\x{26C5}|
\x{26C8}|
\x{26CE}|
\x{26CF}|
\x{26D1}|
\x{26D3}-\x{26D4}|
\x{26E9}-\x{26EA}|
\x{26F0}-\x{26F5}|
\x{26F7}-\x{26FA}|
\x{26FD}|
\x{2702}|
\x{2705}|
\x{2708}-\x{2709}|
\x{270A}-\x{270B}|
\x{270C}-\x{270D}|
\x{270F}|
\x{2712}|
\x{2714}|
\x{2716}|
\x{271D}|
\x{2721}|
\x{2728}|
\x{2733}-\x{2734}|
\x{2744}|
\x{2747}|
\x{274C}|
\x{274E}|
\x{2753}-\x{2755}|
\x{2757}|
\x{2763}-\x{2764}|
\x{2795}-\x{2797}|
\x{27A1}|
\x{27B0}|
\x{27BF}|
\x{2934}-\x{2935}|
\x{2B05}-\x{2B07}|
\x{2B1B}-\x{2B1C}|
\x{2B50}|
\x{2B55}|
\x{3030}|
\x{303D}|
\x{3297}|
\x{3299}|
\x{1F004}|
\x{1F0CF}|
\x{1F170}-\x{1F171}|
\x{1F17E}|
\x{1F17F}|
\x{1F18E}|
\x{1F191}-\x{1F19A}|
\x{1F1E6}-\x{1F1FF}|
\x{1F201}-\x{1F202}|
\x{1F21A}|
\x{1F22F}|
\x{1F232}-\x{1F23A}|
\x{1F250}-\x{1F251}|
\x{1F300}-\x{1F320}|
\x{1F321}|
\x{1F324}-\x{1F32C}|
\x{1F32D}-\x{1F32F}|
\x{1F330}-\x{1F335}|
\x{1F336}|
\x{1F337}-\x{1F37C}|
\x{1F37D}|
\x{1F37E}-\x{1F37F}|
\x{1F380}-\x{1F393}|
\x{1F396}-\x{1F397}|
\x{1F399}-\x{1F39B}|
\x{1F39E}-\x{1F39F}|
\x{1F3A0}-\x{1F3C4}|
\x{1F3C5}|
\x{1F3C6}-\x{1F3CA}|
\x{1F3CB}-\x{1F3CE}|
\x{1F3CF}-\x{1F3D3}|
\x{1F3D4}-\x{1F3DF}|
\x{1F3E0}-\x{1F3F0}|
\x{1F3F3}-\x{1F3F5}|
\x{1F3F7}|
\x{1F3F8}-\x{1F3FF}|
\x{1F400}-\x{1F43E}|
\x{1F43F}|
\x{1F440}|
\x{1F441}|
\x{1F442}-\x{1F4F7}|
\x{1F4F8}|
\x{1F4F9}-\x{1F4FC}|
\x{1F4FD}|
\x{1F4FF}|
\x{1F500}-\x{1F53D}|
\x{1F549}-\x{1F54A}|
\x{1F54B}-\x{1F54E}|
\x{1F550}-\x{1F567}|
\x{1F56F}-\x{1F570}|
\x{1F573}-\x{1F579}|
\x{1F57A}|
\x{1F587}|
\x{1F58A}-\x{1F58D}|
\x{1F590}|
\x{1F595}-\x{1F596}|
\x{1F5A4}|
\x{1F5A5}|
\x{1F5A8}|
\x{1F5B1}-\x{1F5B2}|
\x{1F5BC}|
\x{1F5C2}-\x{1F5C4}|
\x{1F5D1}-\x{1F5D3}|
\x{1F5DC}-\x{1F5DE}|
\x{1F5E1}|
\x{1F5E3}|
\x{1F5E8}|
\x{1F5EF}|
\x{1F5F3}|
\x{1F5FA}|
\x{1F5FB}-\x{1F5FF}|
\x{1F600}|
\x{1F601}-\x{1F610}|
\x{1F611}|
\x{1F612}-\x{1F614}|
\x{1F615}|
\x{1F616}|
\x{1F617}|
\x{1F618}|
\x{1F619}|
\x{1F61A}|
\x{1F61B}|
\x{1F61C}-\x{1F61E}|
\x{1F61F}|
\x{1F620}-\x{1F625}|
\x{1F626}-\x{1F627}|
\x{1F628}-\x{1F62B}|
\x{1F62C}|
\x{1F62D}|
\x{1F62E}-\x{1F62F}|
\x{1F630}-\x{1F633}|
\x{1F634}|
\x{1F635}-\x{1F640}|
\x{1F641}-\x{1F642}|
\x{1F643}-\x{1F644}|
\x{1F645}-\x{1F64F}|
\x{1F680}-\x{1F6C5}|
\x{1F6CB}-\x{1F6CF}|
\x{1F6D0}|
\x{1F6D1}-\x{1F6D2}|
\x{1F6E0}-\x{1F6E5}|
\x{1F6E9}|
\x{1F6EB}-\x{1F6EC}|
\x{1F6F0}|
\x{1F6F3}|
\x{1F6F4}-\x{1F6F6}|
\x{1F910}-\x{1F918}|
\x{1F919}-\x{1F91E}|
\x{1F920}-\x{1F927}|
\x{1F930}|
\x{1F933}-\x{1F93A}|
\x{1F93C}-\x{1F93E}|
\x{1F940}-\x{1F945}|
\x{1F947}-\x{1F94B}|
\x{1F950}-\x{1F95E}|
\x{1F980}-\x{1F984}|
\x{1F985}-\x{1F991}|
\x{1F9C0}
]`
	// reDnmString   = `[a-z][a-z0-9]{2,15}`
	reAmt     = `[[:digit:]]+`
	reDecAmt  = `[[:digit:]]*\.[[:digit:]]+`
	reSpc     = `[[:space:]]*`
	reDnm     = regexp.MustCompile(fmt.Sprintf(`^%s$`, reDnmString))
	reCoin    = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmt, reSpc, reDnmString))
	reDecCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmt, reSpc, reDnmString))
)

// ValidateDenom validates a denomination string returning an error if it is
// invalid.
func ValidateDenom(denom string) error {
	if !reDnm.MatchString(denom) {
		return fmt.Errorf("invalid denom: %s", denom)
	}
	return nil
}

func mustValidateDenom(denom string) {
	if err := ValidateDenom(denom); err != nil {
		panic(err)
	}
}

// ParseCoin parses a cli input for one coin type, returning errors if invalid.
// This returns an error on an empty string as well.
func ParseCoin(coinStr string) (coin Coin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return Coin{}, fmt.Errorf("invalid coin expression: %s", coinStr)
	}

	denomStr, amountStr := matches[2], matches[1]

	amount, ok := NewIntFromString(amountStr)
	if !ok {
		return Coin{}, fmt.Errorf("failed to parse coin amount: %s", amountStr)
	}

	if err := ValidateDenom(denomStr); err != nil {
		return Coin{}, fmt.Errorf("invalid denom cannot contain upper case characters or spaces: %s", err)
	}

	return NewCoin(denomStr, amount), nil
}

// ParseCoins will parse out a list of coins separated by commas.
// If nothing is provided, it returns nil Coins.
// Returned coins are sorted.
func ParseCoins(coinsStr string) (Coins, error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	coinStrs := strings.Split(coinsStr, ",")
	coins := make(Coins, len(coinStrs))
	for i, coinStr := range coinStrs {
		coin, err := ParseCoin(coinStr)
		if err != nil {
			return nil, err
		}

		coins[i] = coin
	}

	// sort coins for determinism
	coins.Sort()

	// validate coins before returning
	if !coins.IsValid() {
		return nil, fmt.Errorf("parseCoins invalid: %#v", coins)
	}

	return coins, nil
}

type findDupDescriptor interface {
	GetDenomByIndex(int) string
	Len() int
}

// findDup works on the assumption that coins is sorted
func findDup(coins findDupDescriptor) int {
	if coins.Len() <= 1 {
		return -1
	}

	prevDenom := coins.GetDenomByIndex(0)
	for i := 1; i < coins.Len(); i++ {
		if coins.GetDenomByIndex(i) == prevDenom {
			return i
		}
		prevDenom = coins.GetDenomByIndex(i)
	}

	return -1
}
