package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// coins with decimal
type DecCoins []DecCoin

// Coins which can have additional decimal points
type DecCoin struct {
	Amount sdk.Dec `json:"amount"`
	Denom  string  `json:"denom"`
}

func NewDecCoin(coin sdk.Coin) DecCoin {
	return DecCoins{
		Amount: sdk.NewDec(coin.Amount),
		Denom:  coin.Denom,
	}
}

func NewDecCoins(coins sdk.Coins) DecCoins {

	dcs := make(DecCoins, len(coins))
	for i, coin := range coins {
		dcs[i] = NewDecCoin(coin)
	}
}
