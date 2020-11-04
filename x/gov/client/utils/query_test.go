package utils_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/rpc/client/mock"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type TxSearchMock struct {
	mock.Client
	txs []tmtypes.Tx
}

func (mock TxSearchMock) TxSearch(ctx context.Context, query string, prove bool, page, perPage *int, orderBy string) (*ctypes.ResultTxSearch, error) {
	if page == nil {
		*page = 0
	}

	if perPage == nil {
		*perPage = 0
	}

	start, end := client.Paginate(len(mock.txs), *page, *perPage, 100)
	if start < 0 || end < 0 {
		// nil result with nil error crashes utils.QueryTxsByEvents
		return &ctypes.ResultTxSearch{}, nil
	}
	txs := mock.txs[start:end]
	rst := &ctypes.ResultTxSearch{Txs: make([]*ctypes.ResultTx, len(txs)), TotalCount: len(txs)}
	for i := range txs {
		rst.Txs[i] = &ctypes.ResultTx{Tx: txs[i]}
	}
	return rst, nil
}

func (mock TxSearchMock) Block(ctx context.Context, height *int64) (*ctypes.ResultBlock, error) {
	// any non nil Block needs to be returned. used to get time value
	return &ctypes.ResultBlock{Block: &tmtypes.Block{}}, nil
}

func newTestCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	types.RegisterLegacyAminoCodec(cdc)
	authtypes.RegisterLegacyAminoCodec(cdc)
	return cdc
}

func TestGetPaginatedVotes(t *testing.T) {
	type testCase struct {
		description string
		page, limit int
		msgs        [][]sdk.Msg
		votes       []types.Vote
	}
	acc1 := make(sdk.AccAddress, 20)
	acc1[0] = 1
	acc2 := make(sdk.AccAddress, 20)
	acc2[0] = 2
	acc1Msgs := []sdk.Msg{
		types.NewMsgVote(acc1, 0, types.OptionYes),
		types.NewMsgVote(acc1, 0, types.OptionYes),
	}
	acc2Msgs := []sdk.Msg{
		types.NewMsgVote(acc2, 0, types.OptionYes),
		types.NewMsgVote(acc2, 0, types.OptionYes),
	}
	for _, tc := range []testCase{
		{
			description: "1MsgPerTxAll",
			page:        1,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
				acc2Msgs[:1],
			},
			votes: []types.Vote{
				types.NewVote(0, acc1, types.OptionYes),
				types.NewVote(0, acc2, types.OptionYes)},
		},
		{
			description: "2MsgPerTx1Chunk",
			page:        1,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs,
				acc2Msgs,
			},
			votes: []types.Vote{
				types.NewVote(0, acc1, types.OptionYes),
				types.NewVote(0, acc1, types.OptionYes)},
		},
		{
			description: "2MsgPerTx2Chunk",
			page:        2,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs,
				acc2Msgs,
			},
			votes: []types.Vote{
				types.NewVote(0, acc2, types.OptionYes),
				types.NewVote(0, acc2, types.OptionYes)},
		},
		{
			description: "IncompleteSearchTx",
			page:        1,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
			},
			votes: []types.Vote{types.NewVote(0, acc1, types.OptionYes)},
		},
		{
			description: "InvalidPage",
			page:        -1,
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
			},
		},
		{
			description: "OutOfBounds",
			page:        2,
			limit:       10,
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
			},
		},
	} {
		tc := tc

		t.Run(tc.description, func(t *testing.T) {
			var (
				marshalled = make([]tmtypes.Tx, len(tc.msgs))
				cdc        = newTestCodec()
			)

			encodingConfig := simapp.MakeTestEncodingConfig()
			cli := TxSearchMock{txs: marshalled}
			clientCtx := client.Context{}.
				WithLegacyAmino(cdc).
				WithClient(cli).
				WithTxConfig(encodingConfig.TxConfig)

			for i := range tc.msgs {
				txBuilder := clientCtx.TxConfig.NewTxBuilder()
				err := txBuilder.SetMsgs(tc.msgs[i]...)
				require.NoError(t, err)

				tx, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
				require.NoError(t, err)
				marshalled[i] = tx
			}

			params := types.NewQueryProposalVotesParams(0, tc.page, tc.limit)
			votesData, err := utils.QueryVotesByTxQuery(clientCtx, params)
			require.NoError(t, err)
			votes := []types.Vote{}
			require.NoError(t, clientCtx.LegacyAmino.UnmarshalJSON(votesData, &votes))
			require.Equal(t, len(tc.votes), len(votes))
			for i := range votes {
				require.Equal(t, tc.votes[i], votes[i])
			}
		})
	}
}
