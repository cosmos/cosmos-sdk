package multisig

import (
	"context"
	"testing"
	"time"

	types "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/defaults/multisig/v1"
)

func setup(t *testing.T, ctx context.Context, ss store.KVStoreService, timefn func() time.Time) *Account {
	t.Helper()
	deps := makeMockDependencies(ss, timefn)

	multisig := NewAccount("multisig")
	_, acc, err := multisig(deps)
	require.NoError(t, err)

	return acc.(*Account)
}

func TestInit(t *testing.T) {
	ctx, ss := newMockContext(t)
	acc := setup(t, ctx, ss, nil)

	testcases := []struct {
		name   string
		msg    *v1.MsgInit
		expErr string
	}{
		{
			"success",
			&v1.MsgInit{
				Config: &v1.Config{
					Threshold:    666,
					Quorum:       400,
					VotingPeriod: 60,
				},
				Members: []*v1.Member{
					{
						Address: "addr1",
						Weight:  500,
					},
					{
						Address: "addr2",
						Weight:  1000,
					},
				},
			},
			"",
		},
		{
			"no members",
			&v1.MsgInit{
				Config: &v1.Config{
					Threshold:    666,
					Quorum:       400,
					VotingPeriod: 60,
				},
				Members: []*v1.Member{},
			},
			"members must be specified",
		},
		{
			"no config",
			&v1.MsgInit{
				Config: nil,
				Members: []*v1.Member{
					{
						Address: "addr1",
						Weight:  500,
					},
				},
			},
			"config must be specified",
		},
		{
			"member weight zero",
			&v1.MsgInit{
				Config: &v1.Config{
					Threshold:    666,
					Quorum:       400,
					VotingPeriod: 60,
				},
				Members: []*v1.Member{
					{
						Address: "addr1",
						Weight:  0,
					},
				},
			},
			"member weight must be greater than zero",
		},
		{
			"threshold is zero",
			&v1.MsgInit{
				Config: &v1.Config{
					Threshold:    0,
					Quorum:       400,
					VotingPeriod: 60,
				},
				Members: []*v1.Member{
					{
						Address: "addr1",
						Weight:  500,
					},
				},
			},
			"threshold, quorum and voting period must be greater than zero",
		},
		{
			"threshold greater than total weight",
			&v1.MsgInit{
				Config: &v1.Config{
					Threshold:    2000,
					Quorum:       400,
					VotingPeriod: 60,
				},
				Members: []*v1.Member{
					{
						Address: "addr1",
						Weight:  500,
					},
					{
						Address: "addr2",
						Weight:  1000,
					},
				},
			},
			"threshold must be less than or equal to the total weight",
		},
		{
			"quorum greater than total weight",
			&v1.MsgInit{
				Config: &v1.Config{
					Threshold:    666,
					Quorum:       2000,
					VotingPeriod: 60,
				},
				Members: []*v1.Member{
					{
						Address: "addr1",
						Weight:  500,
					},
					{
						Address: "addr2",
						Weight:  1000,
					},
				},
			},
			"quorum must be less than or equal to the total weight",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := acc.Init(ctx, tc.msg)
			if tc.expErr != "" {
				require.EqualError(t, err, tc.expErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestUpdateConfig(t *testing.T) {
	// all test cases start from the same initial state
	startAcc := &v1.MsgInit{
		Config: &v1.Config{
			Threshold:    2640,
			Quorum:       2000,
			VotingPeriod: 60,
		},
		Members: []*v1.Member{
			{
				Address: "addr1",
				Weight:  1000,
			},
			{
				Address: "addr2",
				Weight:  1000,
			},
			{
				Address: "addr3",
				Weight:  1000,
			},
			{
				Address: "addr4",
				Weight:  1000,
			},
		},
	}

	testcases := []struct {
		name       string
		msg        *v1.MsgUpdateConfig
		expErr     string
		expCfg     *v1.Config
		expMembers []*v1.Member
	}{
		{
			"change members",
			&v1.MsgUpdateConfig{
				UpdateMembers: []*v1.Member{
					{
						Address: "addr1",
						Weight:  500,
					},
					{
						Address: "addr2",
						Weight:  700,
					},
				},
				Config: &v1.Config{
					Threshold:    666,
					Quorum:       400,
					VotingPeriod: 60,
				},
			},
			"",
			&v1.Config{
				Threshold:    666,
				Quorum:       400,
				VotingPeriod: 60,
			},
			[]*v1.Member{
				{
					Address: "addr1",
					Weight:  500,
				},
				{
					Address: "addr2",
					Weight:  700,
				},
				{
					Address: "addr3",
					Weight:  1000,
				},
				{
					Address: "addr4",
					Weight:  1000,
				},
			},
		},
		{
			"remove member",
			&v1.MsgUpdateConfig{
				UpdateMembers: []*v1.Member{
					{
						Address: "addr1",
						Weight:  0,
					},
				},
				Config: nil,
			},
			"",
			nil,
			[]*v1.Member{
				{
					Address: "addr2",
					Weight:  1000,
				},
				{
					Address: "addr3",
					Weight:  1000,
				},
				{
					Address: "addr4",
					Weight:  1000,
				},
			},
		},
		{
			"add member",
			&v1.MsgUpdateConfig{
				UpdateMembers: []*v1.Member{
					{
						Address: "addr5",
						Weight:  200,
					},
				},
				Config: nil,
			},
			"",
			nil,
			[]*v1.Member{
				{
					Address: "addr1",
					Weight:  1000,
				},
				{
					Address: "addr2",
					Weight:  1000,
				},
				{
					Address: "addr3",
					Weight:  1000,
				},
				{
					Address: "addr4",
					Weight:  1000,
				},
				{
					Address: "addr5",
					Weight:  200,
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, ss := newMockContext(t)
			acc := setup(t, ctx, ss, nil)
			_, err := acc.Init(ctx, startAcc)
			require.NoError(t, err)

			_, err = acc.UpdateConfig(ctx, tc.msg)
			if tc.expErr != "" {
				require.EqualError(t, err, tc.expErr)
				return
			}
			require.NoError(t, err)

			cfg, err := acc.QueryConfig(ctx, &v1.QueryConfig{})
			require.NoError(t, err)

			// if we are not changing the config, we expect the same config as init
			if tc.expCfg == nil {
				require.Equal(t, startAcc.Config, cfg.Config)
			} else {
				require.Equal(t, tc.expCfg, cfg.Config)
			}
			require.Equal(t, tc.expMembers, cfg.Members)
		})
	}
}

func TestProposal_NotPassing(t *testing.T) {
	// all test cases start from the same initial state
	startAcc := &v1.MsgInit{
		Config: &v1.Config{
			Threshold:    2640,
			Quorum:       2000,
			VotingPeriod: 60,
		},
		Members: []*v1.Member{
			{
				Address: "addr1",
				Weight:  1000,
			},
			{
				Address: "addr2",
				Weight:  1000,
			},
			{
				Address: "addr3",
				Weight:  1000,
			},
			{
				Address: "addr4",
				Weight:  1000,
			},
		},
	}

	ctx, ss := accountstd.NewMockContext(
		0, []byte("multisig_acc"), []byte("addr1"), TestFunds, func(ctx context.Context, sender []byte, msg, msgResp ProtoMsg) error {
			return nil
		}, func(ctx context.Context, sender []byte, msg ProtoMsg) (ProtoMsg, error) {
			if _, ok := msg.(*v1.MsgUpdateConfig); ok {
				return &v1.MsgUpdateConfigResponse{}, nil
			}
			return nil, nil
		}, func(ctx context.Context, req, resp ProtoMsg) error {
			return nil
		},
	)

	currentTime := time.Now()

	acc := setup(t, ctx, ss, func() time.Time {
		return currentTime
	})
	_, err := acc.Init(ctx, startAcc)
	require.NoError(t, err)

	msg := &v1.MsgUpdateConfig{
		UpdateMembers: []*v1.Member{
			{
				Address: "addr1",
				Weight:  500,
			},
		},
	}
	anymsg, err := accountstd.PackAny(msg)
	require.NoError(t, err)

	// create a proposal
	createRes, err := acc.CreateProposal(ctx, &v1.MsgCreateProposal{
		Proposal: &v1.Proposal{
			Title:    "test",
			Summary:  "test",
			Messages: []*types.Any{anymsg},
		},
	})
	require.NoError(t, err)

	propId := createRes.ProposalId

	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.NoError(t, err)

	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.ErrorContains(t, err, "voter has already voted, can't change its vote per config")

	ctx = accountstd.SetSender(ctx, []byte("addr2"))
	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.NoError(t, err)

	// try to execute the proposal
	_, err = acc.ExecuteProposal(ctx, &v1.MsgExecuteProposal{
		ProposalId: propId,
	})
	require.ErrorContains(t, err, "voting period has not ended yet")

	// fast forward time
	currentTime = currentTime.Add(61 * time.Second)
	_, err = acc.ExecuteProposal(ctx, &v1.MsgExecuteProposal{
		ProposalId: propId,
	})
	require.NoError(t, err)

	// check proposal status
	prop, err := acc.QueryProposal(ctx, &v1.QueryProposal{
		ProposalId: propId,
	})
	require.NoError(t, err)
	require.Equal(t, v1.ProposalStatus_PROPOSAL_STATUS_REJECTED, prop.Proposal.Status)

	// vote with addr3
	ctx = accountstd.SetSender(ctx, []byte("addr3"))
	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.ErrorContains(t, err, "voting period has ended")
}

func TestProposalPassing(t *testing.T) {
	// all test cases start from the same initial state
	startAcc := &v1.MsgInit{
		Config: &v1.Config{
			Threshold:    2640,
			Quorum:       2000,
			VotingPeriod: 60,
		},
		Members: []*v1.Member{
			{
				Address: "addr1",
				Weight:  1000,
			},
			{
				Address: "addr2",
				Weight:  1000,
			},
			{
				Address: "addr3",
				Weight:  1000,
			},
			{
				Address: "addr4",
				Weight:  1000,
			},
		},
	}

	var acc *Account
	ctx, ss := accountstd.NewMockContext(
		0, []byte("multisig_acc"), []byte("addr1"), TestFunds, func(ctx context.Context, sender []byte, msg, msgResp ProtoMsg) error {
			return nil
		}, func(ctx context.Context, sender []byte, msg ProtoMsg) (ProtoMsg, error) {
			if _, ok := msg.(*v1.MsgUpdateConfig); ok {
				return acc.UpdateConfig(ctx, msg.(*v1.MsgUpdateConfig))
			}
			return nil, nil
		}, func(ctx context.Context, req, resp ProtoMsg) error {
			return nil
		},
	)

	currentTime := time.Now()

	acc = setup(t, ctx, ss, func() time.Time {
		return currentTime
	})
	_, err := acc.Init(ctx, startAcc)
	require.NoError(t, err)

	msg := &v1.MsgUpdateConfig{
		UpdateMembers: []*v1.Member{
			{
				Address: "addr1",
				Weight:  500,
			},
		},
	}
	anymsg, err := accountstd.PackAny(msg)
	require.NoError(t, err)

	// create a proposal
	createRes, err := acc.CreateProposal(ctx, &v1.MsgCreateProposal{
		Proposal: &v1.Proposal{
			Title:    "test",
			Summary:  "test",
			Messages: []*types.Any{anymsg},
		},
	})
	require.NoError(t, err)

	propId := createRes.ProposalId

	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.NoError(t, err)

	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.ErrorContains(t, err, "voter has already voted, can't change its vote per config")

	ctx = accountstd.SetSender(ctx, []byte("addr2"))
	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.NoError(t, err)

	// vote with addr3
	ctx = accountstd.SetSender(ctx, []byte("addr3"))
	_, err = acc.Vote(ctx, &v1.MsgVote{
		ProposalId: propId,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	})
	require.NoError(t, err)

	// fast forward time
	currentTime = currentTime.Add(61 * time.Second)
	_, err = acc.ExecuteProposal(ctx, &v1.MsgExecuteProposal{
		ProposalId: propId,
	})
	require.NoError(t, err)

	// check if addr1's weight changed
	cfg, err := acc.QueryConfig(ctx, &v1.QueryConfig{})
	require.NoError(t, err)
	for _, v := range cfg.Members {
		if v.Address == "addr1" {
			require.Equal(t, uint64(500), v.Weight)
		}
	}
}
