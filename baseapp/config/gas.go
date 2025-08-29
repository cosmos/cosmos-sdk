package config

import sdk "github.com/cosmos/cosmos-sdk/types"

// GasConfig defines the gas configuration parameters for the application.
type GasConfig struct {
	// QueryGasLimit defines the maximum gas limit for queries; if set to 0, queries are unbounded.
	QueryGasLimit uint64

	// MinGasPrices defines the minimum gas prices that a validator is willing to accept
	// for processing a transaction. This is primarily used for DoS and spam prevention.
	MinGasPrices sdk.DecCoins
}
