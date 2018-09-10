package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Coins which can have additional decimal points
type DecCoin struct {
	Denom  string  `json:"denom"`
	Amount sdk.Dec `json:"amount"`
}

func NewDecCoin(coin sdk.Coin) DecCoin {
	return DecCoins{
		Denom:  coin.Denom,
		Amount: sdk.NewDec(coin.Amount),
	}
}

// Adds amounts of two coins with same denom
func (coin DecCoin) Plus(coinB DecCoin) DecCoin {
	if !(coin.Denom == coinB.Denom) {
		return coin
	}
	return Coin{coin.Denom, coin.Amount.Add(coinB.Amount)}
}

// return the decimal coins with trunctated decimals
func (coin DecCoin) TruncateDecimal() sdk.Coin {
	return sdk.NewCoin(coin.Denom, coin.Amount.TruncateInt())
}

//_______________________________________________________________________

// coins with decimal
type DecCoins []DecCoin

func NewDecCoins(coins sdk.Coins) DecCoins {

	dcs := make(DecCoins, len(coins))
	for i, coin := range coins {
		dcs[i] = NewDecCoin(coin)
	}
}

// return the decimal coins with trunctated decimals
func (coins DecCoins) TruncateDecimal() sdk.Coins {
	out := make(DecCoins, len(coins))
	for i, coin := range coins {
		out[i] = coin.TruncateDecimal()
	}
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
