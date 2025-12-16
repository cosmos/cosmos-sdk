package config

import sdk "github.com/cosmos/cosmos-sdk/types"

type GasConfig struct {
	// queryGasLimit defines the maximum gas for queries; unbounded if 0.
	QueryGasLimit uint64

	// The minimum gas prices a validator is willing to accept for processing a
	// transaction. This is mainly used for DoS and spam prevention.
	MinGasPrices sdk.DecCoins
}
