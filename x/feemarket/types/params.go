package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// NewParams instantiates a new EIP-1559 Params object. This params object is utilized
// to implement both the base EIP-1559 fee and AIMD EIP-1559 fee market implementations.
func NewParams(
	window uint64,
	alpha math.LegacyDec,
	beta math.LegacyDec,
	gamma math.LegacyDec,
	delta math.LegacyDec,
	maxBlockSize uint64,
	minBaseGasPrice math.LegacyDec,
	minLearingRate math.LegacyDec,
	maxLearningRate math.LegacyDec,
	feeDenom string,
	enabled bool,
) Params {
	return Params{
		Alpha:               alpha,
		Beta:                beta,
		Gamma:               gamma,
		Delta:               delta,
		MinBaseGasPrice:     minBaseGasPrice,
		MinLearningRate:     minLearingRate,
		MaxLearningRate:     maxLearningRate,
		MaxBlockUtilization: maxBlockSize,
		Window:              window,
		FeeDenom:            feeDenom,
		Enabled:             enabled,
	}
}

// ValidateBasic performs basic validation on the parameters.
func (p *Params) ValidateBasic() error {
	if p.Window == 0 {
		return fmt.Errorf("window cannot be zero")
	}

	if p.Alpha.IsNil() || p.Alpha.IsNegative() {
		return fmt.Errorf("alpha cannot be nil must be between [0, inf)")
	}

	if p.Beta.IsNil() || p.Beta.IsNegative() || p.Beta.GT(math.LegacyOneDec()) {
		return fmt.Errorf("beta cannot be nil and must be between [0, 1]")
	}

	if p.Gamma.IsNil() || p.Gamma.IsNegative() || p.Gamma.GT(math.LegacyMustNewDecFromStr("0.5")) {
		return fmt.Errorf("theta cannot be nil and must be between [0, 0.5]")
	}

	if p.Delta.IsNil() || p.Delta.IsNegative() {
		return fmt.Errorf("delta cannot be nil and must be between [0, inf)")
	}

	if p.MinBaseGasPrice.IsNil() || !p.MinBaseGasPrice.GTE(math.LegacyZeroDec()) {
		return fmt.Errorf("min base gas price cannot be nil and must be greater than or equal to zero")
	}

	if p.MaxLearningRate.IsNil() || p.MinLearningRate.IsNegative() {
		return fmt.Errorf("min learning rate cannot be negative or nil")
	}

	if p.MaxBlockUtilization < 2 {
		return fmt.Errorf("max block utilization cannot be less than 2")
	}

	if p.MaxLearningRate.IsNil() || p.MaxLearningRate.IsNegative() {
		return fmt.Errorf("max learning rate cannot be negative or nil")
	}

	if p.MinLearningRate.GT(p.MaxLearningRate) {
		return fmt.Errorf("min learning rate cannot be greater than max learning rate")
	}

	if p.FeeDenom == "" {
		return fmt.Errorf("fee denom must be set")
	}

	return nil
}

// TargetBlockUtilization returns 0.5 * MaxBlockUtilization.
func (p *Params) TargetBlockUtilization() uint64 {
	return p.MaxBlockUtilization / 2
}
