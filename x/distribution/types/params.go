package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		CommunityTax:        math.LegacyNewDecWithPrec(2, 2), // 2%
		BaseProposerReward:  math.LegacyZeroDec(),            // deprecated
		BonusProposerReward: math.LegacyZeroDec(),            // deprecated
		WithdrawAddrEnabled: true,
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	return validateCommunityTax(p.CommunityTax)
}

func validateCommunityTax(tax math.LegacyDec) error {
	if tax.IsNil() {
		return fmt.Errorf("community tax must be not nil")
	}
	if tax.IsNegative() {
		return fmt.Errorf("community tax must be positive: %s", tax)
	}
	if tax.GT(math.LegacyOneDec()) {
		return fmt.Errorf("community tax too large: %s", tax)
	}

	return nil
}
