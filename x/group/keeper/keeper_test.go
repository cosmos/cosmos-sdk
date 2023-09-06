package keeper_test

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/x/group/module"
	grouptestutil "github.com/cosmos/cosmos-sdk/x/group/testutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

const minExecutionPeriod = 5 * time.Second

type TestSuite struct {
	suite.Suite

	sdkCtx          sdk.Context
	ctx             context.Context
	addrs           []sdk.AccAddress
	groupID         uint64
	groupPolicyAddr sdk.AccAddress
	policy          group.DecisionPolicy
	groupKeeper     keeper.Keeper
	blockTime       time.Time
	bankKeeper      *grouptestutil.MockBankKeeper
	accountKeeper   *grouptestutil.MockAccountKeeper
}

func (s *TestSuite) SetupTest() {
	s.blockTime = cmttime.Now()
	key := storetypes.NewKVStoreKey(group.StoreKey)

	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{}, bank.AppModuleBasic{})
	s.addrs = simtestutil.CreateIncrementalAccounts(6)

	// setup gomock and initialize some globally expected executions
	ctrl := gomock.NewController(s.T())
	s.accountKeeper = grouptestutil.NewMockAccountKeeper(ctrl)
	for i := range s.addrs {
		s.accountKeeper.EXPECT().GetAccount(gomock.Any(), s.addrs[i]).Return(authtypes.NewBaseAccountWithAddress(s.addrs[i])).AnyTimes()
	}
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	s.bankKeeper = grouptestutil.NewMockBankKeeper(ctrl)

	bApp := baseapp.NewBaseApp(
		"group",
		log.NewNopLogger(),
		testCtx.DB,
		encCfg.TxConfig.TxDecoder(),
	)
	bApp.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	banktypes.RegisterMsgServer(bApp.MsgServiceRouter(), s.bankKeeper)

	config := group.DefaultConfig()
	s.groupKeeper = keeper.NewKeeper(key, encCfg.Codec, bApp.MsgServiceRouter(), s.accountKeeper, config)
	s.ctx = testCtx.Ctx.WithBlockTime(s.blockTime)
	s.sdkCtx = sdk.UnwrapSDKContext(s.ctx)

	// Initial group, group policy and balance setup
	members := []group.MemberRequest{
		{Address: s.addrs[4].String(), Weight: "1"}, {Address: s.addrs[1].String(), Weight: "2"},
	}

	s.setNextAccount()

	groupRes, err := s.groupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:   s.addrs[0].String(),
		Members: members,
	})
	s.Require().NoError(err)
	s.groupID = groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		time.Second,
		minExecutionPeriod, // Must wait 5 seconds before executing proposal
	)
	policyReq := &group.MsgCreateGroupPolicy{
		Admin:   s.addrs[0].String(),
		GroupId: s.groupID,
	}
	err = policyReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	s.setNextAccount()

	groupSeq := s.groupKeeper.GetGroupSequence(s.sdkCtx)
	s.Require().Equal(groupSeq, uint64(1))

	policyRes, err := s.groupKeeper.CreateGroupPolicy(s.ctx, policyReq)
	s.Require().NoError(err)

	addrbz, err := address.NewBech32Codec("cosmos").StringToBytes(policyRes.Address)
	s.Require().NoError(err)
	s.policy = policy
	s.groupPolicyAddr = addrbz

	s.bankKeeper.EXPECT().MintCoins(s.sdkCtx, minttypes.ModuleName, sdk.Coins{sdk.NewInt64Coin("test", 100000)}).Return(nil).AnyTimes()
	err = s.bankKeeper.MintCoins(s.sdkCtx, minttypes.ModuleName, sdk.Coins{sdk.NewInt64Coin("test", 100000)})
	s.Require().NoError(err)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.sdkCtx, minttypes.ModuleName, s.groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10000)}).Return(nil).AnyTimes()
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.sdkCtx, minttypes.ModuleName, s.groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10000)})
	s.Require().NoError(err)
}

func (s *TestSuite) setNextAccount() {
	nextAccVal := s.groupKeeper.GetGroupPolicySeq(s.sdkCtx) + 1
	derivationKey := make([]byte, 8)
	binary.BigEndian.PutUint64(derivationKey, nextAccVal)

	ac, err := authtypes.NewModuleCredential(group.ModuleName, []byte{keeper.GroupPolicyTablePrefix}, derivationKey)
	s.Require().NoError(err)

	groupPolicyAcc, err := authtypes.NewBaseAccountWithPubKey(ac)
	s.Require().NoError(err)

	groupPolicyAccBumpAccountNumber, err := authtypes.NewBaseAccountWithPubKey(ac)
	s.Require().NoError(err)
	err = groupPolicyAccBumpAccountNumber.SetAccountNumber(nextAccVal)
	s.Require().NoError(err)

	s.accountKeeper.EXPECT().GetAccount(gomock.Any(), sdk.AccAddress(ac.Address())).Return(nil).AnyTimes()
	s.accountKeeper.EXPECT().NewAccount(gomock.Any(), groupPolicyAcc).Return(groupPolicyAccBumpAccountNumber).AnyTimes()
	s.accountKeeper.EXPECT().SetAccount(gomock.Any(), sdk.AccountI(groupPolicyAccBumpAccountNumber)).Return().AnyTimes()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestProposalsByVPEnd() {
	addrs := s.addrs
	addr2 := addrs[1]

	votingPeriod := s.policy.GetVotingPeriod()
	ctx := s.sdkCtx
	now := time.Now()

	msgSend := &banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}

	proposers := []string{addr2.String()}

	specs := map[string]struct {
		preRun     func(sdkCtx sdk.Context) uint64
		proposalID uint64
		admin      string
		expErrMsg  string
		newCtx     sdk.Context
		tallyRes   group.TallyResult
		expStatus  group.ProposalStatus
	}{
		"tally updated after voting period end": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return submitProposal(sdkCtx, s, []sdk.Msg{msgSend}, proposers)
			},
			admin:     proposers[0],
			newCtx:    ctx.WithBlockTime(now.Add(votingPeriod).Add(time.Hour)),
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_REJECTED,
		},
		"tally within voting period": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return submitProposal(s.ctx, s, []sdk.Msg{msgSend}, proposers)
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_SUBMITTED,
		},
		"tally within voting period (with votes)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return submitProposalAndVote(s.ctx, s, []sdk.Msg{msgSend}, proposers, group.VOTE_OPTION_YES)
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_SUBMITTED,
		},
		"tally after voting period (with votes)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return submitProposalAndVote(s.ctx, s, []sdk.Msg{msgSend}, proposers, group.VOTE_OPTION_YES)
			},
			admin:  proposers[0],
			newCtx: ctx.WithBlockTime(now.Add(votingPeriod).Add(time.Hour)),
			tallyRes: group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				NoWithVetoCount: "0",
				AbstainCount:    "0",
			},
			expStatus: group.PROPOSAL_STATUS_ACCEPTED,
		},
		"tally after voting period (not passing)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				// `s.addrs[4]` has weight 1
				return submitProposalAndVote(s.ctx, s, []sdk.Msg{msgSend}, []string{s.addrs[4].String()}, group.VOTE_OPTION_YES)
			},
			admin:  proposers[0],
			newCtx: ctx.WithBlockTime(now.Add(votingPeriod).Add(time.Hour)),
			tallyRes: group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				NoWithVetoCount: "0",
				AbstainCount:    "0",
			},
			expStatus: group.PROPOSAL_STATUS_REJECTED,
		},
		"tally of withdrawn proposal": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pID := submitProposal(s.ctx, s, []sdk.Msg{msgSend}, proposers)
				_, err := s.groupKeeper.WithdrawProposal(s.ctx, &group.MsgWithdrawProposal{
					ProposalId: pID,
					Address:    proposers[0],
				})

				s.Require().NoError(err)
				return pID
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_WITHDRAWN,
		},
		"tally of withdrawn proposal (with votes)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pID := submitProposalAndVote(s.ctx, s, []sdk.Msg{msgSend}, proposers, group.VOTE_OPTION_YES)
				_, err := s.groupKeeper.WithdrawProposal(s.ctx, &group.MsgWithdrawProposal{
					ProposalId: pID,
					Address:    proposers[0],
				})

				s.Require().NoError(err)
				return pID
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_WITHDRAWN,
		},
	}

	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			pID := spec.preRun(s.sdkCtx)

			err := module.EndBlocker(spec.newCtx, s.groupKeeper)
			s.Require().NoError(err)
			resp, err := s.groupKeeper.Proposal(spec.newCtx, &group.QueryProposalRequest{
				ProposalId: pID,
			})
			s.Require().NoError(err)

			if spec.expErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), spec.expErrMsg)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(resp.GetProposal().FinalTallyResult, spec.tallyRes)
			s.Require().Equal(resp.GetProposal().Status, spec.expStatus)
		})
	}
}

func (s *TestSuite) TestPruneProposals() {
	addrs := s.addrs
	expirationTime := time.Hour * 24 * 15 // 15 days
	groupID := s.groupID
	accountAddr := s.groupPolicyAddr

	msgSend := &banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   addrs[0].String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}

	policyReq := &group.MsgCreateGroupPolicy{
		Admin:   addrs[0].String(),
		GroupId: groupID,
	}

	policy := group.NewThresholdDecisionPolicy("100", time.Microsecond, time.Microsecond)
	err := policyReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)

	s.setNextAccount()

	_, err = s.groupKeeper.CreateGroupPolicy(s.ctx, policyReq)
	s.Require().NoError(err)

	req := &group.MsgSubmitProposal{
		GroupPolicyAddress: accountAddr.String(),
		Proposers:          []string{addrs[1].String()},
	}
	err = req.SetMsgs([]sdk.Msg{msgSend})
	s.Require().NoError(err)
	submittedProposal, err := s.groupKeeper.SubmitProposal(s.ctx, req)
	s.Require().NoError(err)
	queryProposal := group.QueryProposalRequest{ProposalId: submittedProposal.ProposalId}
	prePrune, err := s.groupKeeper.Proposal(s.ctx, &queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(prePrune.Proposal.Id, submittedProposal.ProposalId)
	// Move Forward in time for 15 days, after voting period end + max_execution_period
	s.sdkCtx = s.sdkCtx.WithBlockTime(s.sdkCtx.BlockTime().Add(expirationTime))

	// Prune Expired Proposals
	err = s.groupKeeper.PruneProposals(s.sdkCtx)
	s.Require().NoError(err)
	postPrune, err := s.groupKeeper.Proposal(s.ctx, &queryProposal)
	s.Require().Nil(postPrune)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "load proposal: not found")
}

func submitProposal(
	ctx context.Context, s *TestSuite, msgs []sdk.Msg,
	proposers []string,
) uint64 {
	proposalReq := &group.MsgSubmitProposal{
		GroupPolicyAddress: s.groupPolicyAddr.String(),
		Proposers:          proposers,
	}
	err := proposalReq.SetMsgs(msgs)
	s.Require().NoError(err)

	proposalRes, err := s.groupKeeper.SubmitProposal(ctx, proposalReq)
	s.Require().NoError(err)
	return proposalRes.ProposalId
}

func submitProposalAndVote(
	ctx context.Context, s *TestSuite, msgs []sdk.Msg,
	proposers []string, voteOption group.VoteOption,
) uint64 {
	s.Require().Greater(len(proposers), 0)
	myProposalID := submitProposal(ctx, s, msgs, proposers)

	_, err := s.groupKeeper.Vote(ctx, &group.MsgVote{
		ProposalId: myProposalID,
		Voter:      proposers[0],
		Option:     voteOption,
	})
	s.Require().NoError(err)
	return myProposalID
}

func (s *TestSuite) createGroupAndGroupPolicy(
	admin sdk.AccAddress,
	members []group.MemberRequest,
	policy group.DecisionPolicy,
) (policyAddr string, groupID uint64) {
	groupRes, err := s.groupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:   admin.String(),
		Members: members,
	})
	s.Require().NoError(err)

	groupID = groupRes.GroupId
	groupPolicy := &group.MsgCreateGroupPolicy{
		Admin:   admin.String(),
		GroupId: groupID,
	}

	if policy != nil {
		err = groupPolicy.SetDecisionPolicy(policy)
		s.Require().NoError(err)

		s.setNextAccount()

		groupPolicyRes, err := s.groupKeeper.CreateGroupPolicy(s.ctx, groupPolicy)
		s.Require().NoError(err)
		policyAddr = groupPolicyRes.Address
	}

	return policyAddr, groupID
}

func (s *TestSuite) TestTallyProposalsAtVPEnd() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr2 := addrs[1]
	votingPeriod := 4 * time.Minute
	minExecutionPeriod := votingPeriod + group.DefaultConfig().MaxExecutionPeriod

	groupMsg := &group.MsgCreateGroupWithPolicy{
		Admin: addr1.String(),
		Members: []group.MemberRequest{
			{Address: addr1.String(), Weight: "1"},
			{Address: addr2.String(), Weight: "1"},
		},
	}
	policy := group.NewThresholdDecisionPolicy(
		"1",
		votingPeriod,
		minExecutionPeriod,
	)
	s.Require().NoError(groupMsg.SetDecisionPolicy(policy))

	s.setNextAccount()
	groupRes, err := s.groupKeeper.CreateGroupWithPolicy(s.ctx, groupMsg)
	s.Require().NoError(err)
	accountAddr := groupRes.GetGroupPolicyAddress()
	groupPolicy, err := s.accountKeeper.AddressCodec().StringToBytes(accountAddr)
	s.Require().NoError(err)
	s.Require().NotNil(groupPolicy)

	proposalRes, err := s.groupKeeper.SubmitProposal(s.ctx, &group.MsgSubmitProposal{
		GroupPolicyAddress: accountAddr,
		Proposers:          []string{addr1.String()},
		Messages:           nil,
	})
	s.Require().NoError(err)

	_, err = s.groupKeeper.Vote(s.ctx, &group.MsgVote{
		ProposalId: proposalRes.ProposalId,
		Voter:      addr1.String(),
		Option:     group.VOTE_OPTION_YES,
	})
	s.Require().NoError(err)

	// move forward in time
	ctx := s.sdkCtx.WithBlockTime(s.sdkCtx.BlockTime().Add(votingPeriod + 1))

	result, err := s.groupKeeper.TallyResult(ctx, &group.QueryTallyResultRequest{
		ProposalId: proposalRes.ProposalId,
	})
	s.Require().Equal("1", result.Tally.YesCount)
	s.Require().NoError(err)

	s.Require().NoError(s.groupKeeper.TallyProposalsAtVPEnd(ctx))
	s.NotPanics(func() {
		err := module.EndBlocker(ctx, s.groupKeeper)
		if err != nil {
			panic(err)
		}
	})
}

// TestTallyProposalsAtVPEnd_GroupMemberLeaving test that the node doesn't
// panic if a member leaves after the voting period end.
func (s *TestSuite) TestTallyProposalsAtVPEnd_GroupMemberLeaving() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr3 := addrs[2]
	votingPeriod := 4 * time.Minute
	minExecutionPeriod := votingPeriod + group.DefaultConfig().MaxExecutionPeriod

	groupMsg := &group.MsgCreateGroupWithPolicy{
		Admin: addr1.String(),
		Members: []group.MemberRequest{
			{Address: addr1.String(), Weight: "0.3"},
			{Address: addr2.String(), Weight: "7"},
			{Address: addr3.String(), Weight: "0.6"},
		},
	}
	policy := group.NewThresholdDecisionPolicy(
		"3",
		votingPeriod,
		minExecutionPeriod,
	)
	s.Require().NoError(groupMsg.SetDecisionPolicy(policy))

	s.setNextAccount()
	groupRes, err := s.groupKeeper.CreateGroupWithPolicy(s.ctx, groupMsg)
	s.Require().NoError(err)
	accountAddr := groupRes.GetGroupPolicyAddress()
	groupPolicy, err := sdk.AccAddressFromBech32(accountAddr)
	s.Require().NoError(err)
	s.Require().NotNil(groupPolicy)

	proposalRes, err := s.groupKeeper.SubmitProposal(s.ctx, &group.MsgSubmitProposal{
		GroupPolicyAddress: accountAddr,
		Proposers:          []string{addr1.String()},
		Messages:           nil,
	})
	s.Require().NoError(err)

	// group members vote
	_, err = s.groupKeeper.Vote(s.ctx, &group.MsgVote{
		ProposalId: proposalRes.ProposalId,
		Voter:      addr1.String(),
		Option:     group.VOTE_OPTION_NO,
	})
	s.Require().NoError(err)
	_, err = s.groupKeeper.Vote(s.ctx, &group.MsgVote{
		ProposalId: proposalRes.ProposalId,
		Voter:      addr2.String(),
		Option:     group.VOTE_OPTION_NO,
	})
	s.Require().NoError(err)

	// move forward in time
	ctx := s.sdkCtx.WithBlockTime(s.sdkCtx.BlockTime().Add(votingPeriod + 1))

	// Tally the result. This saves the tally result to state.
	s.Require().NoError(s.groupKeeper.TallyProposalsAtVPEnd(ctx))
	s.NotPanics(func() {
		err := module.EndBlocker(ctx, s.groupKeeper)
		if err != nil {
			panic(err)
		}
	})

	// member 2 (high weight) leaves group.
	_, err = s.groupKeeper.LeaveGroup(ctx, &group.MsgLeaveGroup{
		Address: addr2.String(),
		GroupId: groupRes.GroupId,
	})
	s.Require().NoError(err)

	s.Require().NoError(s.groupKeeper.TallyProposalsAtVPEnd(ctx))
	s.NotPanics(func() {
		err := module.EndBlocker(ctx, s.groupKeeper)
		if err != nil {
			panic(err)
		}
	})
}
