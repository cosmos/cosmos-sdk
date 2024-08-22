package utils_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/client/utils"
	v1 "cosmossdk.io/x/gov/types/v1"

	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

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
			marshaled := make([][]byte, len(tc.msgs))
			fmt.Println(marshaled)
			clientCtx := client.Context{}.
				WithLegacyAmino(encCfg.Amino).
				WithTxConfig(encCfg.TxConfig)

			for i := range tc.msgs {
				txBuilder := clientCtx.TxConfig.NewTxBuilder()
				err := txBuilder.SetMsgs(tc.msgs[i]...)
				require.NoError(t, err)

				tx, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
				require.NoError(t, err)
				marshaled[i] = tx
			}

			cli := clitestutil.MockCometRPC{}.WithTxs(marshaled).WithTxConfig(encCfg.TxConfig)
			clientCtx = clientCtx.WithClient(cli)

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
