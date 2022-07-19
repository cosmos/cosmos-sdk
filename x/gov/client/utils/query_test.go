package utils_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/rpc/client/mock"
	"github.com/tendermint/tendermint/rpc/coretypes"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type TxSearchMock struct {
	txConfig client.TxConfig
	mock.Client
	txs []tmtypes.Tx
}

func (mock TxSearchMock) TxSearch(ctx context.Context, query string, prove bool, page, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	if page == nil {
		*page = 0
	}

	if perPage == nil {
		*perPage = 0
	}

	// Get the `message.action` value from the query.
	messageAction := regexp.MustCompile(`message\.action='(.*)' .*$`)
	msgType := messageAction.FindStringSubmatch(query)[1]

	// Filter only the txs that match the query
	matchingTxs := make([]tmtypes.Tx, 0)
	for _, tx := range mock.txs {
		sdkTx, err := mock.txConfig.TxDecoder()(tx)
		if err != nil {
			return nil, err
		}
		for _, msg := range sdkTx.GetMsgs() {
			if msg.(legacytx.LegacyMsg).Type() == msgType {
				matchingTxs = append(matchingTxs, tx)
				break
			}
		}
	}

	start, end := client.Paginate(len(mock.txs), *page, *perPage, 100)
	if start < 0 || end < 0 {
		// nil result with nil error crashes utils.QueryTxsByEvents
		return &coretypes.ResultTxSearch{}, nil
	}
	if len(matchingTxs) < end {
		return &coretypes.ResultTxSearch{}, nil
	}

	txs := matchingTxs[start:end]
	rst := &coretypes.ResultTxSearch{Txs: make([]*coretypes.ResultTx, len(txs)), TotalCount: len(txs)}
	for i := range txs {
		rst.Txs[i] = &coretypes.ResultTx{Tx: txs[i]}
	}
	return rst, nil
}

func (mock TxSearchMock) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	// any non nil Block needs to be returned. used to get time value
	return &coretypes.ResultBlock{Block: &tmtypes.Block{}}, nil
}

func TestGetPaginatedVotes(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()

	type testCase struct {
		description string
		page, limit int
		msgs        [][]sdk.Msg
		votes       []v1.Vote
	}
	acc1 := make(sdk.AccAddress, 20)
	acc1[0] = 1
	acc2 := make(sdk.AccAddress, 20)
	acc2[0] = 2
	acc1Msgs := []sdk.Msg{
		v1.NewMsgVote(acc1, 0, v1.OptionYes, ""),
		v1.NewMsgVote(acc1, 0, v1.OptionYes, ""),
	}
	acc2Msgs := []sdk.Msg{
		v1.NewMsgVote(acc2, 0, v1.OptionYes, ""),
		v1.NewMsgVoteWeighted(acc2, 0, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
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
			votes: []v1.Vote{
				v1.NewVote(0, acc1, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc2, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
			},
		},
		{
			description: "2MsgPerTx1Chunk",
			page:        1,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs,
				acc2Msgs,
			},
			votes: []v1.Vote{
				v1.NewVote(0, acc1, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc1, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
			},
		},
		{
			description: "2MsgPerTx2Chunk",
			page:        2,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs,
				acc2Msgs,
			},
			votes: []v1.Vote{
				v1.NewVote(0, acc2, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc2, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
			},
		},
		{
			description: "IncompleteSearchTx",
			page:        1,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
			},
			votes: []v1.Vote{v1.NewVote(0, acc1, v1.NewNonSplitVoteOption(v1.OptionYes), "")},
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
			marshalled := make([]tmtypes.Tx, len(tc.msgs))
			cli := TxSearchMock{txs: marshalled, txConfig: encCfg.TxConfig}
			clientCtx := client.Context{}.
				WithLegacyAmino(encCfg.Amino).
				WithClient(cli).
				WithTxConfig(encCfg.TxConfig)

			for i := range tc.msgs {
				txBuilder := clientCtx.TxConfig.NewTxBuilder()
				err := txBuilder.SetMsgs(tc.msgs[i]...)
				require.NoError(t, err)

				tx, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
				require.NoError(t, err)
				marshalled[i] = tx
			}

			params := v1.NewQueryProposalVotesParams(0, tc.page, tc.limit)
			votesData, err := utils.QueryVotesByTxQuery(clientCtx, params)
			require.NoError(t, err)
			votes := []v1.Vote{}
			require.NoError(t, clientCtx.LegacyAmino.UnmarshalJSON(votesData, &votes))
			require.Equal(t, len(tc.votes), len(votes))
			for i := range votes {
				require.Equal(t, tc.votes[i], votes[i])
			}
		})
	}
}
