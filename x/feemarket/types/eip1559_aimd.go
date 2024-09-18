package types

import "cosmossdk.io/math"

// Note: The following constants are the default values for the AIMD EIP-1559
// fee market implementation. This implements an adjustable learning rate
// algorithm that is not present in the base EIP-1559 implementation.

var (
	// DefaultAIMDWindow is the default window size for the sliding window
	// used to calculate the base fee.
	DefaultAIMDWindow uint64 = 8

	// DefaultAIMDAlpha is the default alpha value for the learning
	// rate calculation. This value determines how much we want to additively
	// increase the learning rate when the target block size is exceeded.
	DefaultAIMDAlpha = math.LegacyMustNewDecFromStr("0.025")

	// DefaultAIMDBeta is the default beta value for the learning rate
	// calculation. This value determines how much we want to multiplicatively
	// decrease the learning rate when the target utilization is not met.
	DefaultAIMDBeta = math.LegacyMustNewDecFromStr("0.95")

	// DefaultAIMDGamma is the default threshold for determining whether
	// to increase or decrease the learning rate. In this case, we increase
	// the learning rate if the block utilization within the window is greater
	// than 0.75 or less than 0.25. Otherwise, we multiplicatively decrease
	// the learning rate.
	DefaultAIMDGamma = math.LegacyMustNewDecFromStr("0.25")

	// DefaultAIMDDelta is the default delta value for how much we additively
	// increase or decrease the base fee when the net block utilization within
	// the window is not equal to the target utilization.
	DefaultAIMDDelta = math.LegacyMustNewDecFromStr("0.0")

	// DefaultAIMDMaxBlockSize is the default maximum block utilization.
	// This is the default on Ethereum. This denominated in units of gas
	// consumed in a block.
	DefaultAIMDMaxBlockSize uint64 = 30_000_000

	// DefaultAIMDMinBaseFee is the default minimum base fee.
	DefaultAIMDMinBaseFee = math.LegacyMustNewDecFromStr("1000000000")

	// DefaultAIMDMinLearningRate is the default minimum learning rate.
	DefaultAIMDMinLearningRate = math.LegacyMustNewDecFromStr("0.01")

	// DefaultAIMDMaxLearningRate is the default maximum learning rate.
	DefaultAIMDMaxLearningRate = math.LegacyMustNewDecFromStr("0.50")

	// DefaultAIMDFeeDenom is the Cosmos SDK default bond denom.
	DefaultAIMDFeeDenom = DefaultFeeDenom
)

// DefaultAIMDParams returns a default set of parameters that implements
// the AIMD EIP-1559 fee market implementation. These parameters allow for
// the learning rate to be dynamically adjusted based on the block utilization
// within the window.
func DefaultAIMDParams() Params {
	return NewParams(
		DefaultAIMDWindow,
		DefaultAIMDAlpha,
		DefaultAIMDBeta,
		DefaultAIMDGamma,
		DefaultAIMDDelta,
		DefaultAIMDMaxBlockSize,
		DefaultAIMDMinBaseFee,
		DefaultAIMDMinLearningRate,
		DefaultAIMDMaxLearningRate,
		DefaultAIMDFeeDenom,
		true,
	)
}

// DefaultAIMDState returns the default state for the AIMD EIP-1559 fee market
// implementation. This implementation uses a sliding window to track the
// block utilization and dynamically adjusts the learning rate based on the
// utilization within the window.
func DefaultAIMDState() State {
	return NewState(
		DefaultAIMDWindow,
		DefaultAIMDMinBaseFee,
		DefaultAIMDMinLearningRate,
	)
}

// DefaultAIMDGenesisState returns a default genesis state that implements
// the AIMD EIP-1559 fee market implementation.
func DefaultAIMDGenesisState() *GenesisState {
	return NewGenesisState(DefaultAIMDParams(), DefaultAIMDState())
}
