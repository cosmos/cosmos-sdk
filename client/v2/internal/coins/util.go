package coins

import (
	"errors"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"
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
