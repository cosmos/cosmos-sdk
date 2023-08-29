package cmd

import (
	"strings"

	cmtcfg "github.com/cometbft/cometbft/config"

	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// initCometBFTConfig helps to override default CometBFT Config values.
// return cmtcfg.DefaultConfig if no custom configuration is required for the application.
func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

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
# It can be overwriten by the --gas-adjustment flag in each tx command.
gas-adjustment = {{ .GasConfig.GasAdjustment }}
`)

	return customClientConfigTemplate, customClientConfig
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	// The following code snippet is just for reference.

	// WASMConfig defines configuration for the wasm module.
	type WASMConfig struct {
		// This is the maximum sdk gas (wasm and storage) that we allow for any x/wasm "smart" queries
		QueryGasLimit uint64 `mapstructure:"query-gas-limit"`

		// Address defines the gRPC-web server to listen on
		LruSize uint64 `mapstructure:"lru-size"`
	}

	type CustomAppConfig struct {
		serverconfig.Config `mapstructure:",squash"`

		WASM WASMConfig `mapstructure:"wasm"`
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In simapp, we set the min gas prices to 0.
	srvCfg.MinGasPrices = "0stake"
	// srvCfg.BaseConfig.IAVLDisableFastNode = true // disable fastnode by default

	// Now we set the custom config default values.
	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
		WASM: WASMConfig{
			LruSize:       1,
			QueryGasLimit: 300000,
		},
	}

	// The default SDK app template is defined in serverconfig.DefaultConfigTemplate.
	// We append the custom config template to the default one.
	// And we set the default config to the custom app template.
	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[wasm]
# This is the maximum sdk gas (wasm and storage) that we allow for any x/wasm "smart" queries
query-gas-limit = {{ .WASM.QueryGasLimit }}
# This is the number of wasm vm instances we keep cached in memory for speed-up
# Warning: this is currently unstable and may lead to crashes, best to keep for 0 unless testing locally
lru-size = {{ .WASM.LruSize }}`

	return customAppTemplate, customAppConfig
}
