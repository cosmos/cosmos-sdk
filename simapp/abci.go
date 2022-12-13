package simapp

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

var _ mempool.Mempool = (*NoOpMempool)(nil)

// NoOpMempool defines a no-op mempool. Transactions are completely discarded and
// ignored when BaseApp interacts with the mempool.
type NoOpMempool struct{}

func (m NoOpMempool) Insert(sdk.Context, sdk.Tx) error              { return nil }
func (m NoOpMempool) Select(sdk.Context, [][]byte) mempool.Iterator { return nil }
func (m NoOpMempool) CountTx() int                                  { return 0 }
func (m NoOpMempool) Remove(sdk.Tx) error                           { return nil }

// NoOpPrepareProposal defines a no-op PrepareProposal handler. It will always
// return the transactions sent by the client's request
func NoOpPrepareProposal() sdk.PrepareProposalHandler {
	return func(_ sdk.Context, req abci.RequestPrepareProposal) abci.ResponsePrepareProposal {
		return abci.ResponsePrepareProposal{Txs: req.Txs}
	}
}

// NoOpProcessProposal defines a no-op ProcessProposal Handler. It will always
// return ACCEPT.
func NoOpProcessProposal() sdk.ProcessProposalHandler {
	return func(_ sdk.Context, _ abci.RequestProcessProposal) abci.ResponseProcessProposal {
		return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
	}
}
