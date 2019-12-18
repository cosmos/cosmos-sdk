package utils

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

type txMock struct {
	address sdk.AccAddress
	msgNum  int
}

func (tx txMock) ValidateBasic() sdk.Error {
	return nil
}

func (tx txMock) GetMsgs() (msgs []sdk.Msg) {
	for i := 0; i < tx.msgNum; i++ {
		msgs = append(msgs, types.NewMsgVote(tx.address, 0, types.OptionYes))
	}
	return
}

func makeQuerier(txs []sdk.Tx) TxQuerier {
	return func(cliCtx context.CLIContext, events []string, page, limit int) (*sdk.SearchTxsResult, error) {
		start, end := client.Paginate(len(txs), page, limit, 100)
		if start < 0 || end < 0 {
			return nil, nil
		}
		rst := &sdk.SearchTxsResult{
			TotalCount: len(txs),
			PageNumber: page,
			PageTotal:  len(txs) / limit,
			Limit:      limit,
			Count:      end - start,
		}
		for _, tx := range txs[start:end] {
			rst.Txs = append(rst.Txs, sdk.TxResponse{Tx: tx})
		}
		return rst, nil
	}
}

func TestGetPaginatedVotes(t *testing.T) {
	type testCase struct {
		description string
		page, limit int
		txs         []sdk.Tx
		votes       []types.Vote
	}
	acc1 := make(sdk.AccAddress, 20)
	acc1[0] = 1
	acc2 := make(sdk.AccAddress, 20)
	acc2[0] = 2
	for _, tc := range []testCase{
		{
			description: "1MsgPerTxAll",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 1}, txMock{acc2, 1}},
			votes: []types.Vote{
				types.NewVote(0, acc1, types.OptionYes),
				types.NewVote(0, acc2, types.OptionYes)},
		},
		{
			description: "2MsgPerTx1Chunk",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 2}, txMock{acc2, 2}},
			votes: []types.Vote{
				types.NewVote(0, acc1, types.OptionYes),
				types.NewVote(0, acc1, types.OptionYes)},
		},
		{
			description: "2MsgPerTx2Chunk",
			page:        2,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 2}, txMock{acc2, 2}},
			votes: []types.Vote{
				types.NewVote(0, acc2, types.OptionYes),
				types.NewVote(0, acc2, types.OptionYes)},
		},
		{
			description: "IncompleteSearchTx",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 1}},
			votes:       []types.Vote{types.NewVote(0, acc1, types.OptionYes)},
		},
		{
			description: "IncompleteSearchTx",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 1}},
			votes:       []types.Vote{types.NewVote(0, acc1, types.OptionYes)},
		},
		{
			description: "InvalidPage",
			page:        -1,
			txs:         []sdk.Tx{txMock{acc1, 1}},
		},
		{
			description: "OutOfBounds",
			page:        2,
			limit:       10,
			txs:         []sdk.Tx{txMock{acc1, 1}},
		},
	} {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.CLIContext{}.WithCodec(codec.New())
			params := types.NewQueryProposalVotesParams(0, tc.page, tc.limit)
			votesData, err := QueryVotesByTxQuery(ctx, params, makeQuerier(tc.txs))
			require.NoError(t, err)
			votes := []types.Vote{}
			require.NoError(t, ctx.Codec.UnmarshalJSON(votesData, &votes))
			require.Equal(t, len(tc.votes), len(votes))
			for i := range votes {
				require.Equal(t, tc.votes[i], votes[i])
			}
		})
	}
}
