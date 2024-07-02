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

func (MockCometRPC) BroadcastTxSync(context.Context, cmttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return &coretypes.ResultBroadcastTx{Code: 0}, nil
}

func (m MockCometRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string,
	_ cmtbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}
