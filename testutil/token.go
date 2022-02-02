package testutil

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var DefaultpowerReduction = sdk.NewInt(1_000_000)

// TokensToConsensusPower - convert input tokens to Power
func TokensToConsensusPower(tokens sdk.Int, powerReduction sdk.Int) int64 {
	if tokens.IsNil() || powerReduction.IsNil() || powerReduction.IsZero() {
		return 0
	}

	power := tokens.Quo(powerReduction)

	if power.GT(sdk.NewIntFromUint64(math.MaxInt64)) {
		return 0
	}
	return power.Int64()
}

// TokensFromConsensusPower - convert input power to tokens
func TokensFromConsensusPower(power int64, powerReduction sdk.Int) sdk.Int {
	return sdk.NewInt(power).Mul(powerReduction)
}
