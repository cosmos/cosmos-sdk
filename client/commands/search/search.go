package search

import (
	"github.com/cosmos/cosmos-sdk/client/commands"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// FindTx performs the given search
func FindTx(query string, prove bool) ([]*ctypes.ResultTx, error) {
	client := commands.GetNode()
	// TODO: actually verify these proofs!!!
	return client.TxSearch(query, prove)
}
