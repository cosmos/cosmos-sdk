package utils_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/rpc/client/mock"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/client/utils"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type TxSearchMock struct {
	txConfig client.TxConfig
	mock.Client
	txs []cmttypes.Tx

	// use for filter tx with query conditions
	msgsSet [][]sdk.Msg
}

func (mock TxSearchMock) TxSearch(ctx context.Context, query string, _ bool, page, perPage *int, _ string) (*coretypes.ResultTxSearch, error) {
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
	txs, err := mock.filterTxs(query, start, end)
	if err != nil {
		return nil, err
	}

	rst := &coretypes.ResultTxSearch{Txs: make([]*coretypes.ResultTx, len(txs)), TotalCount: len(txs)}
	for i := range txs {
		rst.Txs[i] = &coretypes.ResultTx{Tx: txs[i]}
	}
	return rst, nil
}

func (mock TxSearchMock) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	// any non nil Block needs to be returned. used to get time value
	return &coretypes.ResultBlock{Block: &cmttypes.Block{}}, nil
}

// mock applying the query string in TxSearch
func (mock TxSearchMock) filterTxs(query string, start, end int) ([]cmttypes.Tx, error) {
	filterTxs := []cmttypes.Tx{}
	proposalIdStr, senderAddr := getQueryAttributes(query)
	if senderAddr != "" {
		proposalId, err := strconv.ParseUint(proposalIdStr, 10, 64)
		if err != nil {
			return nil, err
		}

		for i, msgs := range mock.msgsSet {
			for _, msg := range msgs {
				if voteMsg, ok := msg.(*v1beta1.MsgVote); ok {
					if voteMsg.Voter == senderAddr && voteMsg.ProposalId == proposalId {
						filterTxs = append(filterTxs, mock.txs[i])
						continue
					}
				}

				if voteMsg, ok := msg.(*v1.MsgVote); ok {
					if voteMsg.Voter == senderAddr && voteMsg.ProposalId == proposalId {
						filterTxs = append(filterTxs, mock.txs[i])
						continue
					}
				}

				if voteWeightedMsg, ok := msg.(*v1beta1.MsgVoteWeighted); ok {
					if voteWeightedMsg.Voter == senderAddr && voteWeightedMsg.ProposalId == proposalId {
						filterTxs = append(filterTxs, mock.txs[i])
						continue
					}
				}

				if voteWeightedMsg, ok := msg.(*v1.MsgVoteWeighted); ok {
					if voteWeightedMsg.Voter == senderAddr && voteWeightedMsg.ProposalId == proposalId {
						filterTxs = append(filterTxs, mock.txs[i])
						continue
					}
				}
			}
		}
	} else {
		filterTxs = append(filterTxs, mock.txs...)
	}

	if len(filterTxs) < end {
		return filterTxs, nil
	}

	return filterTxs[start:end], nil
}

// getQueryAttributes extracts value from query string
func getQueryAttributes(q string) (proposalId, senderAddr string) {
	splitStr := strings.Split(q, " OR ")
	if len(splitStr) >= 2 {
		keySender := strings.Trim(splitStr[1], ")")
		senderAddr = strings.Trim(strings.Split(keySender, "=")[1], "'")

		keyProposal := strings.Split(q, " AND ")[0]
		proposalId = strings.Trim(strings.Split(keyProposal, "=")[1], "'")
	} else {
		proposalId = strings.Trim(strings.Split(splitStr[0], "=")[1], "'")
	}

	return proposalId, senderAddr
}

func TestGetPaginatedVotes(t *testing.T) {
	cdcOpts := codectestutil.CodecOptions{}
	encCfg := moduletestutil.MakeTestEncodingConfig(cdcOpts, gov.AppModule{})

	type testCase struct {
		description string
		page, limit int
		msgs        [][]sdk.Msg
		votes       []v1.Vote
	}
	acc1 := make(sdk.AccAddress, 20)
	acc1[0] = 1
	acc1Str, err := cdcOpts.GetAddressCodec().BytesToString(acc1)
	require.NoError(t, err)
	acc2 := make(sdk.AccAddress, 20)
	acc2[0] = 2
	acc2Str, err := cdcOpts.GetAddressCodec().BytesToString(acc2)
	require.NoError(t, err)
	acc1Msgs := []sdk.Msg{
		v1.NewMsgVote(acc1Str, 0, v1.OptionYes, ""),
		v1.NewMsgVote(acc1Str, 0, v1.OptionYes, ""),
		v1.NewMsgDeposit(acc1Str, 0, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))), // should be ignored
	}
	acc2Msgs := []sdk.Msg{
		v1.NewMsgVote(acc2Str, 0, v1.OptionYes, ""),
		v1.NewMsgVoteWeighted(acc2Str, 0, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
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
				v1.NewVote(0, acc1Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc2Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
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
				v1.NewVote(0, acc1Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc1Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
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
				v1.NewVote(0, acc2Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc2Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
			},
		},
		{
			description: "IncompleteSearchTx",
			page:        1,
			limit:       2,
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
			},
			votes: []v1.Vote{v1.NewVote(0, acc1Str, v1.NewNonSplitVoteOption(v1.OptionYes), "")},
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
			marshaled := make([]cmttypes.Tx, len(tc.msgs))
			cli := TxSearchMock{txs: marshaled, txConfig: encCfg.TxConfig}
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
				marshaled[i] = tx
			}

			params := utils.QueryProposalVotesParams{0, tc.page, tc.limit}
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

func TestGetSingleVote(t *testing.T) {
	cdcOpts := codectestutil.CodecOptions{}
	encCfg := moduletestutil.MakeTestEncodingConfig(cdcOpts, gov.AppModule{})

	type testCase struct {
		description string
		msgs        [][]sdk.Msg
		votes       []v1.Vote
		expErr      string
	}
	acc1 := make(sdk.AccAddress, 20)
	acc1[0] = 1
	acc1Str, err := cdcOpts.GetAddressCodec().BytesToString(acc1)
	require.NoError(t, err)
	acc2 := make(sdk.AccAddress, 20)
	acc2[0] = 2
	acc2Str, err := cdcOpts.GetAddressCodec().BytesToString(acc2)
	require.NoError(t, err)
	acc1Msgs := []sdk.Msg{
		v1.NewMsgVote(acc1Str, 0, v1.OptionYes, ""),
		v1.NewMsgVote(acc1Str, 0, v1.OptionYes, ""),
		v1.NewMsgDeposit(acc1Str, 0, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))), // should be ignored
	}
	acc2Msgs := []sdk.Msg{
		v1.NewMsgVote(acc2Str, 0, v1.OptionYes, ""),
		v1.NewMsgVoteWeighted(acc2Str, 0, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
		v1beta1.NewMsgVoteWeighted(acc2Str, 0, v1beta1.NewNonSplitVoteOption(v1beta1.OptionYes)),
	}
	for _, tc := range []testCase{
		{
			description: "no vote found: no msgVote",
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
			},
			votes:  []v1.Vote{},
			expErr: "did not vote on proposalID",
		},
		{
			description: "no vote found: wrong proposal ID",
			msgs: [][]sdk.Msg{
				acc1Msgs[:1],
			},
			votes:  []v1.Vote{},
			expErr: "did not vote on proposalID",
		},
		{
			description: "query 2 voter vote",
			msgs: [][]sdk.Msg{
				acc1Msgs,
				acc2Msgs[:1],
			},
			votes: []v1.Vote{
				v1.NewVote(0, acc1Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc2Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
			},
		},
		{
			description: "query 2 voter vote with v1beta1",
			msgs: [][]sdk.Msg{
				acc1Msgs,
				acc2Msgs[2:],
			},
			votes: []v1.Vote{
				v1.NewVote(0, acc1Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
				v1.NewVote(0, acc2Str, v1.NewNonSplitVoteOption(v1.OptionYes), ""),
			},
		},
	} {
		tc := tc

		t.Run(tc.description, func(t *testing.T) {
			marshaled := make([]cmttypes.Tx, len(tc.msgs))
			cli := TxSearchMock{txs: marshaled, txConfig: encCfg.TxConfig, msgsSet: tc.msgs}
			clientCtx := client.Context{}.
				WithLegacyAmino(encCfg.Amino).
				WithClient(cli).
				WithTxConfig(encCfg.TxConfig).
				WithAddressCodec(cdcOpts.GetAddressCodec()).
				WithCodec(encCfg.Codec)

			for i := range tc.msgs {
				txBuilder := clientCtx.TxConfig.NewTxBuilder()
				err := txBuilder.SetMsgs(tc.msgs[i]...)
				require.NoError(t, err)

				tx, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
				require.NoError(t, err)
				marshaled[i] = tx
			}

			// testing query single vote
			for i, v := range tc.votes {
				accAddr, err := clientCtx.AddressCodec.StringToBytes(v.Voter)
				require.NoError(t, err)
				voteParams := utils.QueryVoteParams{ProposalID: 0, Voter: accAddr}
				voteData, err := utils.QueryVoteByTxQuery(clientCtx, voteParams)
				if tc.expErr != "" {
					require.Error(t, err)
					require.True(t, strings.Contains(err.Error(), tc.expErr))
					continue
				}
				require.NoError(t, err)
				vote := v1.Vote{}
				require.NoError(t, clientCtx.Codec.UnmarshalJSON(voteData, &vote))
				require.Equal(t, v, vote, fmt.Sprintf("vote should be equal at entry %v", i))
			}
		})
	}
}
