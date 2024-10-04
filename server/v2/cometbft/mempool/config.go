package mempool

var DefaultMaxTx = -1

// Config defines the configurations for the SDK built-in app-side mempool implementations.
type Config struct {
	// MaxTxs defines the maximum number of transactions that can be in the mempool.
	MaxTxs int `mapstructure:"max-txs" toml:"max-txs" comment:"max-txs defines the maximum number of transactions that can be in the mempool. A value of 0 indicates an unbounded mempool, a negative value disables the app-side mempool."`
}

// DefaultConfig returns a default configuration for the SDK built-in app-side mempool implementations.
func DefaultConfig() Config {
	return Config{
		MaxTxs: DefaultMaxTx,
	}
}
