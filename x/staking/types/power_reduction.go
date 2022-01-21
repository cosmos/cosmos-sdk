package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func TokensToConsensusPower(tokens sdk.Int, powerReduction sdk.Int) int64 {
	return (tokens.Quo(powerReduction)).Int64()
}

// TokensFromConsensusPower - convert input power to tokens
func TokensFromConsensusPower(power int64, powerReduction sdk.Int) sdk.Int {
	return sdk.NewInt(power).Mul(powerReduction)
}
