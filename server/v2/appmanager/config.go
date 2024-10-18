package appmanager

// Config represents the configuration options for the app manager.
type Config struct {
	ValidateTxGasLimit uint64 `mapstructure:"validate-tx-gas-limit"` // TODO: check how this works on app mempool
	QueryGasLimit      uint64 `mapstructure:"query-gas-limit"`
	SimulationGasLimit uint64 `mapstructure:"simulation-gas-limit"`
}
