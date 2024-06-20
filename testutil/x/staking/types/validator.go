package types

import "cosmossdk.io/math"

func (v Validator) GetTokens() math.Int { return v.Tokens }
