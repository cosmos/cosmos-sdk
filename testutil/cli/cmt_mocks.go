package cli

import (
	"context"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
)

var _ client.CometRPC = (*MockCometRPC)(nil)

type MockCometRPC struct {
	rpcclientmock.Client

	txConfig      client.TxConfig
	txs           []cmttypes.Tx
	responseQuery abci.QueryResponse
}

// NewMockCometRPC returns a mock CometBFT RPC implementation.
// It is used for CLI testing.
func NewMockCometRPC(respQuery abci.QueryResponse) MockCometRPC {
	return MockCometRPC{responseQuery: respQuery}
}

// NewMockCometRPCWithValue returns a mock CometBFT RPC implementation with value only.
// It is used for CLI testing.
func NewMockCometRPCWithValue(bz []byte) MockCometRPC {
	return MockCometRPC{responseQuery: abci.QueryResponse{
		Value: bz,
	}}
}

// accept [][]byte so that module that use this for testing dont have to import comet directly
func (m MockCometRPC) WithTxs(txs [][]byte) MockCometRPC {
	cmtTxs := make([]cmttypes.Tx, len(txs))
	for i, tx := range txs {
		cmtTxs[i] = tx
	}
	m.txs = cmtTxs
	return m
}

func (m MockCometRPC) WithTxConfig(cfg client.TxConfig) MockCometRPC {
	m.txConfig = cfg
	return m
}

func (MockCometRPC) BroadcastTxSync(context.Context, cmttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return &coretypes.ResultBroadcastTx{Code: 0}, nil
}

func (mock MockCometRPC) TxSearch(ctx context.Context, query string, prove bool, page, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	if page == nil {
		*page = 0
	}

	if perPage == nil {
		*perPage = 0
	}

	start, end := client.Paginate(len(mock.txs), *page, *perPage, 100)
	if start < 0 || end < 0 {
		// nil result with nil error crashes utils.QueryTxsByEvents
		return &coretypes.ResultTxSearch{}, nil
	}

	txs := mock.txs[start:end]
	rst := &coretypes.ResultTxSearch{Txs: make([]*coretypes.ResultTx, len(txs)), TotalCount: len(txs)}
	for i := range txs {
		rst.Txs[i] = &coretypes.ResultTx{Tx: txs[i]}
	}
	return rst, nil
}

func (mock MockCometRPC) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	return &coretypes.ResultBlock{Block: &cmttypes.Block{}}, nil
}

func (m MockCometRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string,
	_ cmtbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}
