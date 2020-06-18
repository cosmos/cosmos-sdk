package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultMinGasPrices = ""
)

// BaseConfig defines the server's basic configuration
type BaseConfig struct {
	// The minimum gas prices a validator is willing to accept for processing a
	// transaction. A transaction's fees must meet the minimum of any denomination
	// specified in this config (e.g. 0.25token1;0.0001token2).
	MinGasPrices string `mapstructure:"minimum-gas-prices"`

	Pruning              string `mapstructure:"pruning"`
	PruningKeepEvery     string `mapstructure:"pruning-keep-every"`
	PruningSnapshotEvery string `mapstructure:"pruning-snapshot-every"`

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

// APIConfig defines the API listener configuration.
type APIConfig struct {
	// Enable defines if the API server should be enabled.
	Enable bool `mapstructure:"enable"`

	// Swagger defines if swagger documentation should automatically be registered.
	Swagger bool `mapstructure:"swagger"`

	// EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk)
	EnableUnsafeCORS bool `mapstructure:"enabled-unsafe-cors"`

	// Address defines the API server to listen on
	Address string `mapstructure:"address"`

	// MaxOpenConnections defines the number of maximum open connections
	MaxOpenConnections uint `mapstructure:"max-open-connections"`

	// RPCReadTimeout defines the Tendermint RPC read timeout (in seconds)
	RPCReadTimeout uint `mapstructure:"rpc-read-timeout"`

	// RPCWriteTimeout defines the Tendermint RPC write timeout (in seconds)
	RPCWriteTimeout uint `mapstructure:"rpc-write-timeout"`

	// RPCMaxBodyBytes defines the Tendermint maximum response body (in bytes)
	RPCMaxBodyBytes uint `mapstructure:"rpc-max-body-bytes"`

	// TODO: TLS/Proxy configuration.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/6420
}

// Config defines the server's top level configuration
type Config struct {
	BaseConfig `mapstructure:",squash"`

	// Telemetry defines the application telemetry configuration
	Telemetry telemetry.Config `mapstructure:"telemetry"`
	API       APIConfig        `mapstructure:"api"`
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
			MinGasPrices:         defaultMinGasPrices,
			InterBlockCache:      true,
			Pruning:              store.PruningStrategySyncable,
			PruningKeepEvery:     "0",
			PruningSnapshotEvery: "0",
		},
		Telemetry: telemetry.Config{
			Enabled:      false,
			GlobalLabels: [][]string{},
		},
		API: APIConfig{
			Enable:             false,
			Swagger:            false,
			Address:            "tcp://0.0.0.0:1317",
			MaxOpenConnections: 1000,
			RPCReadTimeout:     10,
			RPCMaxBodyBytes:    1000000,
		},
	}
}

// GetConfig returns a fully parsed Config object.
func GetConfig() Config {
	globalLabelsRaw := viper.Get("telemetry.global-labels").([]interface{})
	globalLabels := make([][]string, 0, len(globalLabelsRaw))
	for _, glr := range globalLabelsRaw {
		labelsRaw := glr.([]interface{})
		if len(labelsRaw) == 2 {
			globalLabels = append(globalLabels, []string{labelsRaw[0].(string), labelsRaw[1].(string)})
		}
	}

	return Config{
		BaseConfig: BaseConfig{
			MinGasPrices:         viper.GetString("minimum-gas-prices"),
			InterBlockCache:      viper.GetBool("inter-block-cache"),
			Pruning:              viper.GetString("pruning"),
			PruningKeepEvery:     viper.GetString("pruning-keep-every"),
			PruningSnapshotEvery: viper.GetString("pruning-snapshot-every"),
			HaltHeight:           viper.GetUint64("halt-height"),
			HaltTime:             viper.GetUint64("halt-time"),
		},
		Telemetry: telemetry.Config{
			ServiceName:             viper.GetString("telemetry.service-name"),
			Enabled:                 viper.GetBool("telemetry.enabled"),
			EnableHostname:          viper.GetBool("telemetry.enable-hostname"),
			EnableHostnameLabel:     viper.GetBool("telemetry.enable-hostname-label"),
			EnableServiceLabel:      viper.GetBool("telemetry.enable-service-label"),
			PrometheusRetentionTime: viper.GetInt64("telemetry.prometheus-retention-time"),
			GlobalLabels:            globalLabels,
		},
		API: APIConfig{
			Enable:             viper.GetBool("api.enable"),
			Address:            viper.GetString("api.address"),
			MaxOpenConnections: viper.GetUint("api.max-open-connections"),
			RPCReadTimeout:     viper.GetUint("api.rpc-read-timeout"),
			RPCWriteTimeout:    viper.GetUint("api.rpc-write-timeout"),
			RPCMaxBodyBytes:    viper.GetUint("api.rpc-max-body-bytes"),
			EnableUnsafeCORS:   viper.GetBool("api.enabled-unsafe-cors"),
		},
	}
}
