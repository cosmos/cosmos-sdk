package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Coins which can have additional decimal points
type DecCoin struct {
	Denom  string  `json:"denom"`
	Amount sdk.Dec `json:"amount"`
}

func NewDecCoin(denom string, amount int64) DecCoin {
	return DecCoin{
		Denom:  denom,
		Amount: sdk.NewDec(amount),
	}
}

func NewDecCoinFromDec(denom string, amount sdk.Dec) DecCoin {
	return DecCoin{
		Denom:  denom,
		Amount: amount,
	}
}

func NewDecCoinFromCoin(coin sdk.Coin) DecCoin {
	return DecCoin{
		Denom:  coin.Denom,
		Amount: sdk.NewDecFromInt(coin.Amount),
	}
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
func (coin DecCoin) TruncateDecimal() (sdk.Coin, DecCoin) {
	truncated := coin.Amount.TruncateInt()
	change := coin.Amount.Sub(sdk.NewDecFromInt(truncated))
	return sdk.NewCoin(coin.Denom, truncated), DecCoin{coin.Denom, change}
}

//_______________________________________________________________________

// coins with decimal
type DecCoins []DecCoin

func NewDecCoins(coins sdk.Coins) DecCoins {
	dcs := make(DecCoins, len(coins))
	for i, coin := range coins {
		dcs[i] = NewDecCoinFromCoin(coin)
	}
	return dcs
}

// return the coins with trunctated decimals, and return the change
func (coins DecCoins) TruncateDecimal() (sdk.Coins, DecCoins) {
	changeSum := DecCoins{}
	out := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		truncated, change := coin.TruncateDecimal()
		out[i] = truncated
		changeSum = changeSum.Plus(DecCoins{change})
	}
	return out, changeSum
}

// Plus combines two sets of coins
// CONTRACT: Plus will never return Coins where one Coin has a 0 amount.
func (coins DecCoins) Plus(coinsB DecCoins) DecCoins {
	sum := ([]DecCoin)(nil)
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
			if coinA.Amount.Add(coinB.Amount).IsZero() {
				// ignore 0 sum coin type
			} else {
				sum = append(sum, coinA.Plus(coinB))
			}
			indexA++
			indexB++
		case 1:
			sum = append(sum, coinB)
			indexB++
		}
	}
}

// Negative returns a set of coins with all amount negative
func (coins DecCoins) Negative() DecCoins {
	res := make([]DecCoin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Neg(),
		})
	}
	return res
}

// Minus subtracts a set of coins from another (adds the inverse)
func (coins DecCoins) Minus(coinsB DecCoins) DecCoins {
	return coins.Plus(coinsB.Negative())
}

// multiply all the coins by a decimal
func (coins DecCoins) MulDec(d sdk.Dec) DecCoins {
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

// divide all the coins by a decimal
func (coins DecCoins) QuoDec(d sdk.Dec) DecCoins {
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

// returns the amount of a denom from deccoins
func (coins DecCoins) AmountOf(denom string) sdk.Dec {
	switch len(coins) {
	case 0:
		return sdk.ZeroDec()
	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return coin.Amount
		}
		return sdk.ZeroDec()
	default:
		midIdx := len(coins) / 2 // binary search
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

// has a negative DecCoin amount
func (coins DecCoins) HasNegative() bool {
	for _, coin := range coins {
		if coin.Amount.IsNegative() {
			return true
		}
	}
	return false
}
