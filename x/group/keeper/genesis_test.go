package keeper_test

import (
	"context"

	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
)

type GenesisTestSuite struct {
	suite.Suite

	app    *simapp.SimApp
	ctx    context.Context
	sdkCtx sdk.Context
	keeper keeper.Keeper
	cdc    *codec.ProtoCodec
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

var (
	memberPub  = secp256k1.GenPrivKey().PubKey()
	accPub     = secp256k1.GenPrivKey().PubKey()
	accAddr    = sdk.AccAddress(accPub.Address())
	memberAddr = sdk.AccAddress(memberPub.Address())
)

func (s *GenesisTestSuite) SetupSuite() {
	checkTx := false
	db := dbm.NewMemDB()
	encCdc := simapp.MakeTestEncodingConfig()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 5, encCdc, simapp.EmptyAppOptions{})

	s.app = app
	s.sdkCtx = app.BaseApp.NewUncachedContext(checkTx, tmproto.Header{})
	s.keeper = app.GroupKeeper
	s.cdc = codec.NewProtoCodec(app.InterfaceRegistry())
	s.ctx = sdk.WrapSDKContext(s.sdkCtx)
}

func (s *GenesisTestSuite) TestInitExportGenesis() {
	sdkCtx := s.sdkCtx
	ctx := s.ctx
	cdc := s.cdc

	submittedAt := time.Now().UTC()
	timeout := submittedAt.Add(time.Second * 1).UTC()

	groupPolicy := &group.GroupPolicyInfo{
		Address:  accAddr.String(),
		GroupId:  1,
		Admin:    accAddr.String(),
		Version:  1,
		Metadata: []byte("policy metadata"),
	}
	err := groupPolicy.SetDecisionPolicy(&group.ThresholdDecisionPolicy{
		Threshold: "1",
		Timeout:   time.Second,
	})
	s.Require().NoError(err)

	proposal := &group.Proposal{
		ProposalId:         1,
		Address:            accAddr.String(),
		Metadata:           []byte("proposal metadata"),
		GroupVersion:       1,
		GroupPolicyVersion: 1,
		Proposers: []string{
			memberAddr.String(),
		},
		SubmittedAt: submittedAt,
		Status:      group.ProposalStatusClosed,
		Result:      group.ProposalResultAccepted,
		VoteState: group.Tally{
			YesCount:     "1",
			NoCount:      "0",
			AbstainCount: "0",
			VetoCount:    "0",
		},
		Timeout:        timeout,
		ExecutorResult: group.ProposalExecutorResultSuccess,
	}
	err = proposal.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: accAddr.String(),
		ToAddress:   memberAddr.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}})
	s.Require().NoError(err)

	genesisState := &group.GenesisState{
		GroupSeq:              2,
		Groups:                []*group.GroupInfo{{GroupId: 1, Admin: accAddr.String(), Metadata: []byte("1"), Version: 1, TotalWeight: "1"}, {GroupId: 2, Admin: accAddr.String(), Metadata: []byte("2"), Version: 2, TotalWeight: "2"}},
		GroupMembers:          []*group.GroupMember{{GroupId: 1, Member: &group.Member{Address: memberAddr.String(), Weight: "1", Metadata: []byte("member metadata")}}, {GroupId: 2, Member: &group.Member{Address: memberAddr.String(), Weight: "2", Metadata: []byte("member metadata")}}},
		GroupPolicySeq: 1,
		GroupPolicies:         []*group.GroupPolicyInfo{groupPolicy},
		ProposalSeq:           1,
		Proposals:             []*group.Proposal{proposal},
		Votes:                 []*group.Vote{{ProposalId: proposal.ProposalId, Voter: memberAddr.String(), SubmittedAt: submittedAt, Choice: group.Choice_CHOICE_YES}},
	}
	genesisBytes, err := cdc.MarshalJSON(genesisState)
	s.Require().NoError(err)

	genesisData := map[string]json.RawMessage{
		group.ModuleName: genesisBytes,
	}

	s.keeper.InitGenesis(sdkCtx, cdc, genesisData[group.ModuleName])

	for i, g := range genesisState.Groups {
		res, err := s.keeper.GroupInfo(ctx, &group.QueryGroupInfoRequest{
			GroupId: g.GroupId,
		})
		s.Require().NoError(err)
		s.Require().Equal(g, res.Info)

		membersRes, err := s.keeper.GroupMembers(ctx, &group.QueryGroupMembersRequest{
			GroupId: g.GroupId,
		})
		s.Require().NoError(err)
		s.Require().Equal(len(membersRes.Members), 1)
		s.Require().Equal(membersRes.Members[0], genesisState.GroupMembers[i])
	}

	for _, g := range genesisState.GroupPolicies {
		res, err := s.keeper.GroupPolicyInfo(ctx, &group.QueryGroupPolicyInfoRequest{
			Address: g.Address,
		})
		s.Require().NoError(err)
		s.assertGroupPoliciesEqual(g, res.Info)
	}

	for _, g := range genesisState.Proposals {
		res, err := s.keeper.Proposal(ctx, &group.QueryProposalRequest{
			ProposalId: g.ProposalId,
		})
		s.Require().NoError(err)
		s.assertProposalsEqual(g, res.Proposal)

		votesRes, err := s.keeper.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{
			ProposalId: g.ProposalId,
		})
		s.Require().NoError(err)
		s.Require().Equal(len(votesRes.Votes), 1)
		s.Require().Equal(votesRes.Votes[0], genesisState.Votes[0])
	}

	exported := s.keeper.ExportGenesis(sdkCtx, cdc)
	bz, err := cdc.MarshalJSON(exported)
	s.Require().NoError(err)

	var exportedGenesisState group.GenesisState
	err = cdc.UnmarshalJSON(bz, &exportedGenesisState)
	s.Require().NoError(err)

	s.Require().Equal(genesisState.Groups, exportedGenesisState.Groups)
	s.Require().Equal(genesisState.GroupMembers, exportedGenesisState.GroupMembers)

	s.Require().Equal(len(genesisState.GroupPolicies), len(exportedGenesisState.GroupPolicies))
	for i, g := range genesisState.GroupPolicies {
		res := exportedGenesisState.GroupPolicies[i]
		s.Require().NoError(err)
		s.assertGroupPoliciesEqual(g, res)
	}

	s.Require().Equal(len(genesisState.Proposals), len(exportedGenesisState.Proposals))
	for i, g := range genesisState.Proposals {
		res := exportedGenesisState.Proposals[i]
		s.Require().NoError(err)
		s.assertProposalsEqual(g, res)
	}
	s.Require().Equal(genesisState.Votes, exportedGenesisState.Votes)

	s.Require().Equal(genesisState.GroupSeq, exportedGenesisState.GroupSeq)
	s.Require().Equal(genesisState.GroupPolicySeq, exportedGenesisState.GroupPolicySeq)
	s.Require().Equal(genesisState.ProposalSeq, exportedGenesisState.ProposalSeq)

}

func (s *GenesisTestSuite) assertGroupPoliciesEqual(g *group.GroupPolicyInfo, other *group.GroupPolicyInfo) {
	require := s.Require()
	require.Equal(g.Address, other.Address)
	require.Equal(g.GroupId, other.GroupId)
	require.Equal(g.Admin, other.Admin)
	require.Equal(g.Metadata, other.Metadata)
	require.Equal(g.Version, other.Version)
	require.Equal(g.GetDecisionPolicy(), other.GetDecisionPolicy())
}

func (s *GenesisTestSuite) assertProposalsEqual(g *group.Proposal, other *group.Proposal) {
	require := s.Require()
	require.Equal(g.ProposalId, other.ProposalId)
	require.Equal(g.Address, other.Address)
	require.Equal(g.Metadata, other.Metadata)
	require.Equal(g.Proposers, other.Proposers)
	require.Equal(g.SubmittedAt, other.SubmittedAt)
	require.Equal(g.GroupVersion, other.GroupVersion)
	require.Equal(g.GroupPolicyVersion, other.GroupPolicyVersion)
	require.Equal(g.Status, other.Status)
	require.Equal(g.Result, other.Result)
	require.Equal(g.VoteState, other.VoteState)
	require.Equal(g.Timeout, other.Timeout)
	require.Equal(g.ExecutorResult, other.ExecutorResult)
	require.Equal(g.GetMsgs(), other.GetMsgs())
}
