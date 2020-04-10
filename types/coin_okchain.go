package types

import (
	"fmt"
	"math/big"
)

//-----------------------------------------------------------------------------
// Coin: alias to DecCoin
type Coin = DecCoin

//-----------------------------------------------------------------------------
// Coins: alias to DecCoins
type Coins = DecCoins

//-----------------------------------------------------------------------------

func NewCoin(denom string, amount Int) DecCoin {
	return NewDecCoin(denom, amount)
}

func NewInt64Coin(denom string, amount int64) DecCoin {
	return NewCoin(denom, NewInt(amount))
}

func NewDecCoinsFromDec(denom string, amount Dec) DecCoins {
	return DecCoins{NewDecCoinFromDec(denom, amount)}
}

func (dec DecCoin) ToCoins() Coins {
	return NewCoins(dec)
}

// Round a decimal with precision, perform bankers rounding (gaussian rounding)
func (d Dec) RoundDecimal(precision int64) Dec {
	precisionMul := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(precision), nil))
	return newDecFromInt(d.MulInt(precisionMul).RoundInt()).QuoInt(precisionMul)
}

func ValidateDenom(denom string) error {
	return validateDenom(denom)
}

func MustParseCoins(denom, amount string) Coins {
	coins, err := ParseCoins(amount + denom)
	if err != nil {
		panic(err)
	}
	return coins
}

func validate(denom string, amount Int) error {
	if err := validateDenom(denom); err != nil {
		return err
	}

	if amount.LT(ZeroInt()) {
		return fmt.Errorf("negative coin amount: %v", amount)
	}

	return nil
}

func GetSystemFee() Coin {
	return NewDecCoinFromDec(DefaultBondDenom, NewDecWithPrec(125, 4))
}

func ZeroFee() Coin {
	return NewCoin(DefaultBondDenom, ZeroInt())
}
