package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Note: We use the same default values as Ethereum for the EIP-1559
// fee market implementation. These parameters do not implement the
// AIMD learning rate adjustment algorithm.

var (
	// DefaultWindow is the default window size for the sliding window
	// used to calculate the base fee. In the base EIP-1559 implementation,
	// only the previous block is considered.
	DefaultWindow uint64 = 1

	// DefaultAlpha is not used in the base EIP-1559 implementation.
	DefaultAlpha = math.LegacyMustNewDecFromStr("0.0")

	// DefaultBeta is not used in the base EIP-1559 implementation.
	DefaultBeta = math.LegacyMustNewDecFromStr("1.0")

	// DefaultGamma is not used in the base EIP-1559 implementation.
	DefaultGamma = math.LegacyMustNewDecFromStr("0.0")

	// DefaultDelta is not used in the base EIP-1559 implementation.
	DefaultDelta = math.LegacyMustNewDecFromStr("0.0")

	// DefaultMaxBlockUtilization is the default maximum block utilization. This is the default
	// on Ethereum. This denominated in units of gas consumed in a block.
	DefaultMaxBlockUtilization uint64 = 30_000_000

	// DefaultMinBaseGasPrice is the default minimum base fee.
	DefaultMinBaseGasPrice = math.LegacyOneDec()

	// DefaultMinLearningRate is not used in the base EIP-1559 implementation.
	DefaultMinLearningRate = math.LegacyMustNewDecFromStr("0.125")

	// DefaultMaxLearningRate is not used in the base EIP-1559 implementation.
	DefaultMaxLearningRate = math.LegacyMustNewDecFromStr("0.125")

	// DefaultFeeDenom is the Cosmos SDK default bond denom.
	DefaultFeeDenom = sdk.DefaultBondDenom
)

// DefaultParams returns a default set of parameters that implements
// the EIP-1559 fee market implementation without the AIMD learning
// rate adjustment algorithm.
func DefaultParams() Params {
	return NewParams(
		DefaultWindow,
		DefaultAlpha,
		DefaultBeta,
		DefaultGamma,
		DefaultDelta,
		DefaultMaxBlockUtilization,
		DefaultMinBaseGasPrice,
		DefaultMinLearningRate,
		DefaultMaxLearningRate,
		DefaultFeeDenom,
		true,
	)
}

// DefaultState returns the default state for the EIP-1559 fee market
// implementation without the AIMD learning rate adjustment algorithm.
func DefaultState() State {
	return NewState(
		DefaultWindow,
		DefaultMinBaseGasPrice,
		DefaultMinLearningRate,
	)
}

// DefaultGenesisState returns a default genesis state that implements
// the EIP-1559 fee market implementation without the AIMD learning
// rate adjustment algorithm.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), DefaultState())
}
