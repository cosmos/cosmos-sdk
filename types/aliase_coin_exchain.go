package types

import (
	"fmt"
	"math/big"
	"regexp"
)

type Coin = DecCoin
type Coins = DecCoins

type SysCoin = DecCoin
type SysCoins = DecCoins


var (
	reDnmString = fmt.Sprintf(`[a-z][a-z0-9]{0,9}(\-[a-f0-9]{3})?`)
	reDecAmt    = `[[:digit:]]*\.?[[:digit:]]+`

	rePoolTokenDnmString = fmt.Sprintf(`(ammswap_)[a-z][a-z0-9]{0,9}(\-[a-f0-9]{3})?_[a-z][a-z0-9]{0,9}(\-[a-f0-9]{3})?`)
	rePoolTokenDnm       = regexp.MustCompile(fmt.Sprintf(`^%s$`, rePoolTokenDnmString))
	reDecCoinPoolToken   = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmt, reSpc, rePoolTokenDnmString))

	reDecCoin = &EnvRegexp{
		regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmt, reSpc, reDnmString)),
	}
)

var (
	ParseCoin = ParseDecCoin
	ParseCoins = ParseDecCoins
)

func NewCoin(denom string, amount interface{}) DecCoin {
	switch amount := amount.(type) {
	case Int:
		return NewDecCoin(denom, amount)
	case Dec:
		return NewDecCoinFromDec(denom, amount)
	default:
		panic("Invalid amount")
	}
}

func NewDecCoinsFromDec(denom string, amount Dec) DecCoins {
	return DecCoins{NewDecCoinFromDec(denom, amount)}
}

func (dec DecCoin) ToCoins() Coins {
	return NewCoins(dec)
}

func newDecFromInt(i Int) Dec {
	return newDecFromIntWithPrec(i, 0)
}

func newDecFromIntWithPrec(i Int, prec int64) Dec {
	return Dec{
		new(big.Int).Mul(i.BigInt(), precisionMultiplier(prec)),
	}
}
// Round a decimal with precision, perform bankers rounding (gaussian rounding)
func (d Dec) RoundDecimal(precision int64) Dec {
	precisionMul := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(precision), nil))
	return newDecFromInt(d.MulInt(precisionMul).RoundInt()).QuoInt(precisionMul)
}


func MustParseCoins(denom, amount string) Coins {
	coins, err := ParseCoins(amount + denom)
	if err != nil {
		panic(err)
	}
	return coins
}

type EnvRegexp struct {
	*regexp.Regexp
}

func (r *EnvRegexp) FindStringSubmatch(coinStr string) []string {
	matches := r.Regexp.FindStringSubmatch(coinStr)
	if matches != nil {
		return matches
	}
	return reDecCoinPoolToken.FindStringSubmatch(coinStr)
}

func GetSystemFee() Coin {
	return NewDecCoinFromDec(DefaultBondDenom, NewDecWithPrec(125, 4))
}

func ZeroFee() Coin {
	return NewCoin(DefaultBondDenom, ZeroInt())
}

// ValidateDenom validates a denomination string returning an error if it is
// invalid.
func ValidateDenom(denom string) error {
	if !reDnm.MatchString(denom) && !rePoolTokenDnm.MatchString(denom) {
		return fmt.Errorf("invalid denom: %s", denom)
	}
	return nil
}

func (coins DecCoins) Add2(coinsB DecCoins) DecCoins {
	return coins.safeAdd(coinsB)
}

