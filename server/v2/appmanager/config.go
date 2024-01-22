package appmanager

type Config struct {
	ValidateTxGasLimit uint64 `mapstructure:"validate-tx-gas-limit"`
	queryGasLimit      uint64 `mapstructure:"query-gas-limit"`
	simulationGasLimit uint64 `mapstructure:"simulation-gas-limit"`
}
