package simulation_test

import (
	"math/big"

	sdkmath "github.com/cosmos/cosmos-sdk/math/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	sdk.DefaultPowerReduction = sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
}
