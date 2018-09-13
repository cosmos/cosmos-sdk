package rest

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// contains checks if the a given query contains one of the tx types
func contains(stringSlice []string, txType string) bool {
	for _, word := range stringSlice {
		if word == txType {
			return true
		}
	}
	return false
}

// queries staking txs
func queryTxs(node rpcclient.Client, cdc *codec.Codec, tag string, delegatorAddr string) ([]tx.Info, error) {
	page := 0
	perPage := 100
	prove := false
	query := fmt.Sprintf("%s='%s' AND %s='%s'", tags.Action, tag, tags.Delegator, delegatorAddr)
	res, err := node.TxSearch(query, prove, page, perPage)
	if err != nil {
		return nil, err
	}

	return tx.FormatTxResults(cdc, res.Txs)
}
