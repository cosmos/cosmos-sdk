package coin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Coin hold some amount of one currency
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

// String provides a human-readable representation of a coin
func (coin Coin) String() string {
	return fmt.Sprintf("%v%v", coin.Amount, coin.Denom)
}

// IsZero returns if this represents no money
func (coin Coin) IsZero() bool {
	return coin.Amount == 0
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (coin Coin) IsGTE(other Coin) bool {
	return (coin.Denom == other.Denom) &&
		(coin.Amount >= other.Amount)
}

//----------------------------------------
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

// IsValid asserts the Coins are sorted, and don't have 0 amounts
func (coins Coins) IsValid() bool {
	switch len(coins) {
	case 0:
		return true
	case 1:
		return coins[0].Amount != 0
	default:
		lowDenom := coins[0].Denom
		for _, coin := range coins[1:] {
			if coin.Denom <= lowDenom {
				return false
			}
			if coin.Amount == 0 {
				return false
			}
			// we compare each coin against the last denom
			lowDenom = coin.Denom
		}
		return true
	}
}

// Plus combines to sets of coins
//
// TODO: handle empty coins!
// Currently appends an empty coin ...
func (coins Coins) Plus(coinsB Coins) Coins {
	sum := []Coin{}
	indexA, indexB := 0, 0
	lenA, lenB := len(coins), len(coinsB)
	for {
		if indexA == lenA {
			if indexB == lenB {
				return sum
			}
			return append(sum, coinsB[indexB:]...)
		} else if indexB == lenB {
			return append(sum, coins[indexA:]...)
		}
		coinA, coinB := coins[indexA], coinsB[indexB]
		switch strings.Compare(coinA.Denom, coinB.Denom) {
		case -1:
			sum = append(sum, coinA)
			indexA++
		case 0:
			if coinA.Amount+coinB.Amount == 0 {
				// ignore 0 sum coin type
			} else {
				sum = append(sum, Coin{
					Denom:  coinA.Denom,
					Amount: coinA.Amount + coinB.Amount,
				})
			}
			indexA++
			indexB++
		case 1:
			sum = append(sum, coinB)
			indexB++
		}
	}
	return sum
}

// Negative returns a set of coins with all amount negative
func (coins Coins) Negative() Coins {
	res := make([]Coin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, Coin{
			Denom:  coin.Denom,
			Amount: -coin.Amount,
		})
	}
	return res
}

// Minus subtracts a set of coins from another (adds the inverse)
func (coins Coins) Minus(coinsB Coins) Coins {
	return coins.Plus(coinsB.Negative())
}

// IsGTE returns True iff coins is NonNegative(), and for every
// currency in coinsB, the currency is present at an equal or greater
// amount in coinsB
func (coins Coins) IsGTE(coinsB Coins) bool {
	diff := coins.Minus(coinsB)
	if len(diff) == 0 {
		return true
	}
	return diff.IsNonnegative()
}

// IsZero returns true if there are no coins
func (coins Coins) IsZero() bool {
	return len(coins) == 0
}

// IsEqual returns true if the two sets of Coins have the same value
func (coins Coins) IsEqual(coinsB Coins) bool {
	if len(coins) != len(coinsB) {
		return false
	}
	for i := 0; i < len(coins); i++ {
		if coins[i] != coinsB[i] {
			return false
		}
	}
	return true
}

// IsPositive returns true if there is at least one coin, and all
// currencies have a positive value
func (coins Coins) IsPositive() bool {
	if len(coins) == 0 {
		return false
	}
	for _, coinAmount := range coins {
		if coinAmount.Amount <= 0 {
			return false
		}
	}
	return true
}

// IsNonnegative returns true if there is no currency with a negative value
// (even no coins is true here)
func (coins Coins) IsNonnegative() bool {
	if len(coins) == 0 {
		return true
	}
	for _, coinAmount := range coins {
		if coinAmount.Amount < 0 {
			return false
		}
	}
	return true
}

//----------------------------------------
// Sort interface

//nolint
func (coins Coins) Len() int           { return len(coins) }
func (coins Coins) Less(i, j int) bool { return coins[i].Denom < coins[j].Denom }
func (coins Coins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

var _ sort.Interface = Coins{}

// Sort is a helper function to sort the set of coins inplace
func (coins Coins) Sort() { sort.Sort(coins) }
