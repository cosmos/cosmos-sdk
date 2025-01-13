package coin

import "math/big"

type Coin struct {
	Denom  string
	Amount big.Int
}

type Coins []Coin
