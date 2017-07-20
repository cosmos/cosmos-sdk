package commands

import (
	"errors"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/modules/base"
)

//nolint
const (
	FlagExpires = "expires"
)

// ChainWrapper wraps a tx with an chain info and optional expiration
type ChainWrapper struct{}

var _ txcmd.Wrapper = ChainWrapper{}

// Wrap will wrap the tx with a ChainTx from the standard flags
func (ChainWrapper) Wrap(tx basecoin.Tx) (res basecoin.Tx, err error) {
	expires := viper.GetInt64(FlagExpires)
	chain := commands.GetChainID()
	if chain == "" {
		return res, errors.New("No chain-id provided")
	}
	res = base.NewChainTx(chain, uint64(expires), tx)
	return
}

// Register adds the sequence flags to the cli
func (ChainWrapper) Register(fs *pflag.FlagSet) {
	fs.Uint64(FlagExpires, 0, "Block height at which this tx expires")
}
