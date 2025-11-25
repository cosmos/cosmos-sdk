// package config

import sdk "github.com/cosmos/cosmos-sdk/types"

// GasConfig defines the configuration parameters related to gas limits and pricing.
type GasConfig struct {
	// QueryGasLimit defines the maximum gas allowed for a query.
	// It is unbounded if the value is set to 0.
	QueryGasLimit uint64 

	// MinGasPrices represents the minimum gas prices that a validator is willing
	// to accept for processing a transaction. This is primarily used for
	// denial-of-service (DoS) and spam prevention.
	MinGasPrices sdk.DecCoins
}
