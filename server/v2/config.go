package serverv2

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// ServerConfig defines configuration for the main server.
type ServerConfig struct {
	MinGasPrices string `mapstructure:"minimum-gas-prices" toml:"minimum-gas-prices" comment:"minimum-gas-prices defines the price which a validator is willing to accept for processing a transaction. A transaction's fees must meet the minimum of any denomination specified in this config (e.g. 0.25token1;0.0001token2)."`
}

// ValidateBasic returns an error if min-gas-prices field is empty in Config. Otherwise, it returns nil.
func (c ServerConfig) ValidateBasic() error {
	if c.MinGasPrices == "" {
		return fmt.Errorf("error in app.toml: set minimum-gas-prices in app.toml or flag or env variable")
	}

	return nil
}

// DefaultMainServerConfig returns the default config of main server component
func DefaultMainServerConfig() ServerConfig {
	return ServerConfig{}
}

// OverwriteDefaultConfig overwrites main server config with given config
func (s *Server[T]) OverwriteDefaultConfig(cfg ServerConfig) {
	s.config = cfg
}

// ReadConfig returns a viper instance of the config file
func ReadConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigName("config")
	v.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %s: %w", configPath, err)
	}

	v.SetConfigName("app")
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to merge configuration: %w", err)
	}

	v.WatchConfig()

	return v, nil
}

// UnmarshalSubConfig unmarshals the given subconfig from the viper instance.
// It unmarshals the config, env, flags into the target struct.
// Use this instead of viper.Sub because viper does not unmarshal flags.
func UnmarshalSubConfig(v *viper.Viper, subName string, target any) error {
	var sub any
	for k, val := range v.AllSettings() {
		if k == subName {
			sub = val
			break
		}
	}

	// Create a new decoder with custom decoding options
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		Result:           target,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	// Decode the sub-configuration
	if err := decoder.Decode(sub); err != nil {
		return fmt.Errorf("failed to decode sub-configuration: %w", err)
	}

	return nil
}
