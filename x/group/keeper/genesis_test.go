package keeper_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	proto "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
)

type GenesisTestSuite struct {
	suite.Suite

	app              *simapp.SimApp
	ctx              context.Context
	genesisCtx       sdk.Context
	keeper           keeper.Keeper
	cdc              *codec.ProtoCodec
	addrs            []sdk.AccAddress
	groupAccountAddr sdk.AccAddress
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupSuite() {
	checkTx := false
	app := simapp.Setup(s.T(), checkTx)
	interfaceRegistry := types.NewInterfaceRegistry()
	group.RegisterInterfaces(interfaceRegistry)

	s.app = app
	s.genesisCtx = app.BaseApp.NewContext(checkTx, tmproto.Header{})
	s.keeper = app.GroupKeeper
	s.cdc = codec.NewProtoCodec(app.InterfaceRegistry())
	s.ctx = sdk.WrapSDKContext(s.genesisCtx)
	s.addrs = simapp.AddTestAddrsIncremental(app, s.genesisCtx, 3, sdk.NewInt(30000000))
	s.groupAccountAddr = s.addrs[0]
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.genesisCtx, s.groupAccountAddr, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))

	s.T().Parallel()
}

func (s *GenesisTestSuite) TestInitExportGenesis() {
	genesisCtx := s.genesisCtx
	ctx := s.ctx
	cdc := s.cdc

	now := time.Now()
	psubmittedAt, err := proto.TimestampProto(now)
	s.Require().NoError(err)
	submittedAt, err := proto.TimestampFromProto(psubmittedAt)
	s.Require().NoError(err)
	ptimeout, err := proto.TimestampProto(now.Add(time.Second * 1))
	s.Require().NoError(err)
	timeout, err := proto.TimestampFromProto(ptimeout)
	s.Require().NoError(err)

	groupAccount := &group.GroupAccountInfo{
		Address:  s.groupAccountAddr.String(),
		GroupId:  1,
		Admin:    s.addrs[0].String(),
		Version:  1,
		Metadata: []byte("account metadata"),
	}
	err = groupAccount.SetDecisionPolicy(&group.ThresholdDecisionPolicy{
		Threshold: "1",
		Timeout:   time.Second,
	})
	s.Require().NoError(err)

	proposal := &group.Proposal{
		ProposalId:          1,
		Address:             s.groupAccountAddr.String(),
		Metadata:            []byte("proposal metadata"),
		GroupVersion:        1,
		GroupAccountVersion: 1,
		Proposers: []string{
			s.addrs[0].String(),
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
		FromAddress: s.groupAccountAddr.String(),
		ToAddress:   s.addrs[1].String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}})
	s.Require().NoError(err)

	genesisState := &group.GenesisState{
		GroupSeq:        2,
		Groups:          []*group.GroupInfo{{GroupId: 1, Admin: s.addrs[0].String(), Metadata: []byte("1"), Version: 1, TotalWeight: "1"}, {GroupId: 2, Admin: s.addrs[1].String(), Metadata: []byte("2"), Version: 2, TotalWeight: "2"}},
		GroupMembers:    []*group.GroupMember{{GroupId: 1, Member: &group.Member{Address: s.addrs[0].String(), Weight: "1", Metadata: []byte("member metadata")}}, {GroupId: 2, Member: &group.Member{Address: s.addrs[0].String(), Weight: "2", Metadata: []byte("member metadata")}}},
		GroupAccountSeq: 1,
		GroupAccounts:   []*group.GroupAccountInfo{groupAccount},
		ProposalSeq:     1,
		Proposals:       []*group.Proposal{proposal},
		Votes:           []*group.Vote{{ProposalId: proposal.ProposalId, Voter: s.addrs[0].String(), SubmittedAt: submittedAt, Choice: group.Choice_CHOICE_YES}},
	}

	genesisBytes, err := cdc.MarshalJSON(genesisState)
	s.Require().NoError(err)

	genesisData := map[string]json.RawMessage{
		group.ModuleName: genesisBytes,
	}

	s.keeper.InitGenesis(genesisCtx, cdc, genesisData[group.ModuleName])

	for i, g := range genesisState.Groups {
		_, err := s.keeper.GroupInfo(ctx, &group.QueryGroupInfo{
			GroupId: g.GroupId,
		})
		s.Require().NoError(err)

		membersRes, err := s.keeper.GroupMembers(ctx, &group.QueryGroupMembers{
			GroupId: g.GroupId,
		})
		s.Require().NoError(err)
		s.Require().Equal(len(membersRes.Members), 1)
		s.Require().Equal(membersRes.Members[0], genesisState.GroupMembers[i])
	}

	for _, g := range genesisState.GroupAccounts {
		res, err := s.keeper.GroupAccountInfo(ctx, &group.QueryGroupAccountInfo{
			Address: g.Address,
		})
		s.Require().NoError(err)
		s.assertGroupAccountsEqual(g, res.Info)
	}

	for _, g := range genesisState.Proposals {
		res, err := s.keeper.Proposal(ctx, &group.QueryProposal{
			ProposalId: g.ProposalId,
		})
		s.Require().NoError(err)
		s.assertProposalsEqual(g, res.Proposal)

		votesRes, err := s.keeper.VotesByProposal(ctx, &group.QueryVotesByProposal{
			ProposalId: g.ProposalId,
		})
		s.Require().NoError(err)
		s.Require().Equal(len(votesRes.Votes), 1)
		s.Require().Equal(votesRes.Votes[0], genesisState.Votes[0])
	}

	exported := s.keeper.ExportGenesis(genesisCtx, cdc)
	bz, err := cdc.MarshalJSON(exported)
	s.Require().NoError(err)

	var exportedGenesisState group.GenesisState
	err = cdc.UnmarshalJSON(bz, &exportedGenesisState)
	s.Require().NoError(err)

	s.Require().Equal(genesisState.Groups, exportedGenesisState.Groups)
	s.Require().Equal(genesisState.GroupMembers, exportedGenesisState.GroupMembers)

	s.Require().Equal(len(genesisState.GroupAccounts), len(exportedGenesisState.GroupAccounts))
	for i, g := range genesisState.GroupAccounts {
		res := exportedGenesisState.GroupAccounts[i]
		s.Require().NoError(err)
		s.assertGroupAccountsEqual(g, res)
	}

	s.Require().Equal(len(genesisState.Proposals), len(exportedGenesisState.Proposals))
	for i, g := range genesisState.Proposals {
		res := exportedGenesisState.Proposals[i]
		s.Require().NoError(err)
		s.assertProposalsEqual(g, res)
	}
	s.Require().Equal(genesisState.Votes, exportedGenesisState.Votes)

	s.Require().Equal(genesisState.GroupSeq, exportedGenesisState.GroupSeq)
	s.Require().Equal(genesisState.GroupAccountSeq, exportedGenesisState.GroupAccountSeq)
	s.Require().Equal(genesisState.ProposalSeq, exportedGenesisState.ProposalSeq)

}

func (s *GenesisTestSuite) assertGroupAccountsEqual(g *group.GroupAccountInfo, other *group.GroupAccountInfo) {
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
	require.Equal(g.GroupAccountVersion, other.GroupAccountVersion)
	require.Equal(g.Status, other.Status)
	require.Equal(g.Result, other.Result)
	require.Equal(g.VoteState, other.VoteState)
	require.Equal(g.Timeout, other.Timeout)
	require.Equal(g.ExecutorResult, other.ExecutorResult)
	require.Equal(g.GetMsgs(), other.GetMsgs())
}
