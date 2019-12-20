package types

import "fmt"

func CoinFromCoinE(e CoinE) (coin Coin, err error) {
	amt, err := e.Amount()
	if err != nil {
		return coin, err
	}
	x, ok := NewIntFromString(amt)
	if !ok {
		return coin, fmt.Errorf("can't parse Int")
	}
	denom, err := e.Denom()
	if err != nil {
		return coin, err
	}
	return Coin{
		Amount: x,
		Denom:  denom,
	}, nil
}

func CoinsFromCoinEList(list CoinE_List) (coins Coins, err error) {
	panic("TODO")
}
