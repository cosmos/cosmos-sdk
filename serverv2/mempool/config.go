package mempool

// Config defines the configurations for the SDK built-in app-side mempool
// implementations.
type Config struct {
	// MaxTxs defines the behavior of the mempool. A negative value indicates
	// the mempool is disabled entirely, zero indicates that the mempool is
	// unbounded in how many txs it may contain, and a positive value indicates
	// the maximum amount of txs it may contain.
	MaxTxs int `mapstructure:"max-txs"`
}
