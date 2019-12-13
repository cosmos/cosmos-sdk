package utils

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
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
		msgs = append(msgs, types.MsgVote{
			Voter: tx.address,
		})
	}
	return
}

func makeQuerier(txs []sdk.Tx) txQuerier {
	return func(cliCtx context.CLIContext, events []string, page, limit int) (*sdk.SearchTxsResult, error) {
		start, end := client.Paginate(len(txs), page, limit, 100)
		if start < 0 || end < 0 {
			return nil, errors.New("unexpected error in test")
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
		err         error
	}
	acc1 := sdk.AccAddress{1}
	acc2 := sdk.AccAddress{2}
	for _, tc := range []testCase{
		{
			description: "1MsgPerTxAll",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 1}, txMock{acc2, 1}},
			votes:       []types.Vote{{Voter: acc1}, {Voter: acc2}},
		},
		{
			description: "2MsgPerTx1Chunk",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 2}, txMock{acc2, 2}},
			votes:       []types.Vote{{Voter: acc1}, {Voter: acc1}},
		},
		{
			description: "2MsgPerTx2Chunk",
			page:        2,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 2}, txMock{acc2, 2}},
			votes:       []types.Vote{{Voter: acc2}, {Voter: acc2}},
		},
		{
			description: "IncompleteSearchTx",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 1}},
			votes:       []types.Vote{{Voter: acc1}},
		},
		{
			description: "IncompleteSearchTx",
			page:        1,
			limit:       2,
			txs:         []sdk.Tx{txMock{acc1, 1}},
			votes:       []types.Vote{{Voter: acc1}},
		},
		{
			description: "InvalidPage",
			page:        -1,
			txs:         []sdk.Tx{txMock{acc1, 1}},
			err:         fmt.Errorf("invalid page (%d) or range is out of bounds (limit %d)", -1, 0),
		},
		{
			description: "OutOfBounds",
			page:        2,
			limit:       10,
			txs:         []sdk.Tx{txMock{acc1, 1}},
			err:         fmt.Errorf("invalid page (%d) or range is out of bounds (limit %d)", 2, 10),
		},
	} {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			params := types.NewQueryProposalVotesParams(0, tc.page, tc.limit)
			votes, err := getPaginatedVotes(context.CLIContext{}, params, makeQuerier(tc.txs))
			if tc.err != nil {
				require.NotNil(t, err)
				require.EqualError(t, tc.err, err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, len(tc.votes), len(votes))
			for i := range votes {
				require.Equal(t, tc.votes[i], votes[i])
			}
		})
	}
}
