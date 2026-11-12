package tx_test

import (
	"context"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// blockClient serves a single block with the given txs.
type blockClient struct {
	client.MockClient
	txs []cmttypes.Tx
}

func (c blockClient) Block(_ context.Context, height *int64) (*coretypes.ResultBlock, error) {
	return &coretypes.ResultBlock{Block: cmttypes.MakeBlock(*height, c.txs, nil, nil)}, nil
}

func TestGetBlockWithTxsPagination(t *testing.T) {
	interfaceRegistry := testutil.CodecOptions{}.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	txConfig := tx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), tx.DefaultSignModes)

	memos := []string{"tx0", "tx1", "tx2"}
	blockTxs := make([]cmttypes.Tx, len(memos))
	for i, memo := range memos {
		builder := txConfig.NewTxBuilder()
		builder.SetMemo(memo)
		bz, err := txConfig.TxEncoder()(builder.GetTx())
		require.NoError(t, err)
		blockTxs[i] = bz
	}

	clientCtx := client.Context{}.WithTxConfig(txConfig).WithClient(blockClient{txs: blockTxs})
	srv := tx.NewTxServer(clientCtx, nil, interfaceRegistry)
	ctx := sdk.NewContext(nil, cmtproto.Header{Height: 1}, false, log.NewNopLogger())

	tests := []struct {
		name       string
		pagination *query.PageRequest
		want       []string
	}{
		{"no pagination", nil, []string{"tx0", "tx1", "tx2"}},
		{"explicit limit", &query.PageRequest{Limit: 2}, []string{"tx0", "tx1"}},
		{"offset without limit", &query.PageRequest{Offset: 1}, []string{"tx1", "tx2"}},
		{"count_total without limit", &query.PageRequest{CountTotal: true}, []string{"tx0", "tx1", "tx2"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := srv.GetBlockWithTxs(ctx, &txtypes.GetBlockWithTxsRequest{Height: 1, Pagination: tc.pagination})
			require.NoError(t, err)

			got := make([]string, len(res.Txs))
			for i, theTx := range res.Txs {
				got[i] = theTx.Body.Memo
			}
			require.Equal(t, tc.want, got)
			require.Equal(t, uint64(len(memos)), res.Pagination.Total)
		})
	}
}
