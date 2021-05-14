package config

import (
	"fmt"
	"strings"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultMinGasPrices = "0.0000000001" + sdk.DefaultBondDenom
)

// BaseConfig defines the server's basic configuration
type BaseConfig struct {
	// The minimum gas prices a validator is willing to accept for processing a
	// transaction. A transaction's fees must meet the minimum of any denomination
	// specified in this config (e.g. 0.25token1;0.0001token2).
	MinGasPrices string `mapstructure:"minimum-gas-prices"`

	Pruning           string `mapstructure:"pruning"`
	PruningKeepRecent string `mapstructure:"pruning-keep-recent"`
	PruningKeepEvery  string `mapstructure:"pruning-keep-every"`
	PruningInterval   string `mapstructure:"pruning-interval"`

	// HaltHeight contains a non-zero block height at which a node will gracefully
	// halt and shutdown that can be used to assist upgrades and testing.
	//
	// Note: Commitment of state will be attempted on the corresponding block.
	HaltHeight uint64 `mapstructure:"halt-height"`

	// HaltTime contains a non-zero minimum block time (in Unix seconds) at which
	// a node will gracefully halt and shutdown that can be used to assist
	// upgrades and testing.
	//
	// Note: Commitment of state will be attempted on the corresponding block.
	HaltTime uint64 `mapstructure:"halt-time"`

	// InterBlockCache enables inter-block caching.
	InterBlockCache bool `mapstructure:"inter-block-cache"`
}

// Config defines the server's top level configuration
type Config struct {
	BaseConfig    `mapstructure:",squash"`
	BackendConfig *BackendConfig `mapstructure:"backend"`
	StreamConfig  *StreamConfig  `mapstructure:"stream"`
}

// SetMinGasPrices sets the validator's minimum gas prices.
func (c *Config) SetMinGasPrices(gasPrices sdk.DecCoins) {
	c.MinGasPrices = gasPrices.String()
}

// GetMinGasPrices returns the validator's minimum gas prices based on the set
// configuration.
func (c *Config) GetMinGasPrices() sdk.DecCoins {
	if c.MinGasPrices == "" {
		return sdk.DecCoins{}
	}

	gasPricesStr := strings.Split(c.MinGasPrices, ";")
	gasPrices := make(sdk.DecCoins, len(gasPricesStr))

	for i, s := range gasPricesStr {
		gasPrice, err := sdk.ParseDecCoin(s)
		if err != nil {
			panic(fmt.Errorf("failed to parse minimum gas price coin (%s): %s", s, err))
		}

		gasPrices[i] = gasPrice
	}

	return gasPrices
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			MinGasPrices:      defaultMinGasPrices,
			InterBlockCache:   true,
			Pruning:           storetypes.PruningOptionDefault,
			PruningKeepRecent: "0",
			PruningKeepEvery:  "0",
			PruningInterval:   "0",
		},
		BackendConfig: DefaultBackendConfig(),
		StreamConfig:  DefaultStreamConfig(),
	}
}
