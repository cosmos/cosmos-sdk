package types

import (
	"fmt"

	"cosmossdk.io/math"
)

const (
	// DefaultNakamotoBonusPeriodEpochIdentifier represents default nakamoto bonus period epoch identifier
	DefaultNakamotoBonusPeriodEpochIdentifier = "week"
)

var (
	DefaultNakamotoBonusStep               = math.LegacyNewDecWithPrec(1, 2)
	DefaultNakamotoBonusMinimumCoefficient = math.LegacyNewDecWithPrec(3, 2)
	DefaultNakamotoBonusMaximumCoefficient = math.LegacyOneDec()
)

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		CommunityTax:        math.LegacyNewDecWithPrec(2, 2), // 2%
		WithdrawAddrEnabled: true,
		NakamotoBonus: NakamotoBonus{
			Enabled:               true,
			Step:                  DefaultNakamotoBonusStep,
			PeriodEpochIdentifier: DefaultNakamotoBonusPeriodEpochIdentifier,
			MinimumCoefficient:    DefaultNakamotoBonusMinimumCoefficient,
			MaximumCoefficient:    DefaultNakamotoBonusMaximumCoefficient,
		},
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	if err := validateCommunityTax(p.CommunityTax); err != nil {
		return err
	}
	return validateNakamotoBonus(p.NakamotoBonus)
}

func validateCommunityTax(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	switch {
	case v.IsNil():
		return fmt.Errorf("community tax must be not nil")
	case v.IsNegative():
		return fmt.Errorf("community tax must be positive: %s", v)
	case v.GT(math.LegacyOneDec()):
		return fmt.Errorf("community tax too large: %s", v)
	}
	return nil
}

func validateWithdrawAddrEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateNakamotoBonus(v NakamotoBonus) error {
	if v.PeriodEpochIdentifier == "" {
		return fmt.Errorf("nakamoto bonus period epoch identifier must not be empty")
	}

	switch {
	case v.Step.IsNil():
		return fmt.Errorf("nakamoto bonus step must be not nil")
	case v.Step.IsNegative() || v.Step.IsZero():
		return fmt.Errorf("nakamoto bonus step must be positive: %v", v.Step)
	case v.Step.GT(math.LegacyOneDec()):
		return fmt.Errorf("nakamoto bonus step too large: %v", v.Step)
	}

	switch {
	case v.MinimumCoefficient.IsNil():
		return fmt.Errorf("nakamoto bonus minimum must be not nil")
	case v.MinimumCoefficient.IsNegative() || v.MinimumCoefficient.IsZero():
		return fmt.Errorf("nakamoto bonus minimum must be positive: %v", v.MinimumCoefficient)
	case v.MinimumCoefficient.GT(math.LegacyOneDec()):
		return fmt.Errorf("nakamoto bonus minimum too large: %v", v.MinimumCoefficient)
	}

	switch {
	case v.MaximumCoefficient.IsNil():
		return fmt.Errorf("nakamoto bonus maximum must be not nil")
	case v.MaximumCoefficient.IsNegative() || v.MaximumCoefficient.IsZero():
		return fmt.Errorf("nakamoto bonus maximum must be positive: %v", v.Step)
	case v.MaximumCoefficient.GT(math.LegacyOneDec()):
		return fmt.Errorf("nakamoto bonus maximum too large: %v", v.MaximumCoefficient)
	}

	if v.MinimumCoefficient.GT(v.MaximumCoefficient) {
		return fmt.Errorf("nakamoto bonus minimum (%v) can't be greater than maximum (%v)", v.MinimumCoefficient, v.MaximumCoefficient)
	}

	return nil
}
