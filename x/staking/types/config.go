package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Config is a config struct used for intialising the staking module to avoid using globals.
type Config struct {
	// PowerReduction is the amount of staking tokens required for 1 unit of consensus-engine power
	PowerReduction sdk.Int
}

func DefaultConfig() Config {
	return Config{
		PowerReduction: sdk.NewIntFromUint64(1000000),
	}
}
