package mempool

const ConfigTemplateMempool = `
###############################################################################
###                         Mempool                                         ###
###############################################################################

[mempool]
# Setting max-txs to 0 will allow for a unbounded amount of transactions in the mempool.
# Setting max_txs to negative 1 (-1) will disable transactions from being inserted into the mempool.
# Setting max_txs to a positive number (> 0) will limit the number of transactions in the mempool, by the specified amount.
#
# Note, this configuration only applies to SDK built-in app-side mempool
# implementations.
max-txs = {{ .Mempool.MaxTxs }}
`

// Config defines the configurations for the SDK built-in app-side mempool
// implementations.
type Config struct {
	// MaxTxs defines the behavior of the mempool. A negative value indicates
	// the mempool is disabled entirely, zero indicates that the mempool is
	// unbounded in how many txs it may contain, and a positive value indicates
	// the maximum amount of txs it may contain.
	MaxTxs int `mapstructure:"max-txs"`
}
