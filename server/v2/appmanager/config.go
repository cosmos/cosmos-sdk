package appmanager

type Config struct {
	ValidateTxGasLimit uint64 `mapstructure:"validate-tx-gas-limit"`
	QueryGasLimit      uint64 `mapstructure:"query-gas-limit"`
	SimulationGasLimit uint64 `mapstructure:"simulation-gas-limit"`
}
