package rosetta

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SdkClient interface {
	GetAuthAccount(ctx context.Context, address string, height int64) (types.AccountResponse, error)
	GetTx(ctx context.Context, hash string) (sdk.TxResponse, error)
	PostTx(ctx context.Context, bytes []byte) (sdk.TxResponse, error)
	GetNodeInfo(ctx context.Context) (rpc.NodeInfoResponse, error)
}

type TendermintClient interface {
	NetInfo() (tendermint.NetInfoResponse, error)
	Block(height uint64) (tendermint.BlockResponse, error)
	BlockByHash(hash string) (tendermint.BlockResponse, error)
	Status() (tendermint.StatusResponse, error)
	UnconfirmedTxs() (tendermint.UnconfirmedTxsResponse, error)
	Tx(hash string) (tendermint.TxResponse, error)
	TxSearch(query string) (tendermint.TxSearchResponse, error)
}
