package coins

import (
	"errors"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ withAmount = &base.Coin{}
	_ withAmount = &base.DecCoin{}
)

type withAmount interface {
	GetAmount() string
}

// IsZero check if given coins are zero.
func IsZero[T withAmount](coins []T) (bool, error) {
	for _, coin := range coins {
		amount, ok := math.NewIntFromString(coin.GetAmount())
		if !ok {
			return false, errors.New("invalid coin amount")
		}
		if !amount.IsZero() {
			return false, nil
		}
	}
	return true, nil
}

func ParseDecCoins(coins string) ([]*base.DecCoin, error) {
	parsedGasPrices, err := sdk.ParseDecCoins(coins) // TODO: do it here to avoid sdk dependency
	if err != nil {
		return nil, err
	}

	finalGasPrices := make([]*base.DecCoin, len(parsedGasPrices))
	for i, coin := range parsedGasPrices {
		finalGasPrices[i] = &base.DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		}
	}
	return finalGasPrices, nil
}

func ParseCoinsNormalized(coins string) ([]*base.Coin, error) {
	parsedFees, err := sdk.ParseCoinsNormalized(coins) // TODO: do it here to avoid sdk dependency
	if err != nil {
		return nil, err
	}

	finalFees := make([]*base.Coin, len(parsedFees))
	for i, coin := range parsedFees {
		finalFees[i] = &base.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		}
	}

	return finalFees, nil
}
