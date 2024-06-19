package cmd

import (
	"strings"

	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// initAppConfig helps to override default client config template and configs.
// return "", nil if no custom configuration is required for the application.
func initClientConfig() (string, interface{}) {
	type GasConfig struct {
		GasAdjustment float64 `mapstructure:"gas-adjustment"`
	}

	type CustomClientConfig struct {
		clientconfig.Config `mapstructure:",squash"`

		GasConfig GasConfig `mapstructure:"gas"`
	}

	// Optionally allow the chain developer to overwrite the SDK's default client config.
	clientCfg := clientconfig.DefaultConfig()

	// The SDK's default keyring backend is set to "os".
	// This is more secure than "test" and is the recommended value.
	//
	// In simapp, we set the default keyring backend to test, as SimApp is meant
	// to be an example and testing application.
	clientCfg.KeyringBackend = keyring.BackendTest

	// Now we set the custom config default values.
	customClientConfig := CustomClientConfig{
		Config: *clientCfg,
		GasConfig: GasConfig{
			GasAdjustment: 1.5,
		},
	}

	// The default SDK app template is defined in serverconfig.DefaultConfigTemplate.
	// We append the custom config template to the default one.
	// And we set the default config to the custom app template.
	customClientConfigTemplate := clientconfig.DefaultClientConfigTemplate + strings.TrimSpace(`
# This is default the gas adjustment factor used in tx commands.
# It can be overwritten by the --gas-adjustment flag in each tx command.
gas-adjustment = {{ .GasConfig.GasAdjustment }}
`)

	return customClientConfigTemplate, customClientConfig
}
