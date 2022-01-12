package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
)

type invariantTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	cdc       *codec.ProtoCodec
	key       *storetypes.KVStoreKey
	blockTime time.Time
}

func TestInvariantTestSuite(t *testing.T) {
	suite.Run(t, new(invariantTestSuite))
}

func (s *invariantTestSuite) SetupSuite() {
	interfaceRegistry := types.NewInterfaceRegistry()
	group.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	key := sdk.NewKVStoreKey(group.ModuleName)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	_ = cms.LoadLatestVersion()
	sdkCtx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	s.ctx = sdkCtx
	s.cdc = cdc
	s.key = key

}

func (s *invariantTestSuite) TestTallyVotesInvariant() {
	sdkCtx, _ := s.ctx.CacheContext()
	curCtx, cdc, key := sdkCtx, s.cdc, s.key
	prevCtx, _ := curCtx.CacheContext()
	prevCtx = prevCtx.WithBlockHeight(curCtx.BlockHeight() - 1)

	// Proposal Table
	proposalTable, err := orm.NewAutoUInt64Table([2]byte{keeper.ProposalTablePrefix}, keeper.ProposalTableSeqPrefix, &group.Proposal{}, cdc)
	s.Require().NoError(err)

	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	specs := map[string]struct {
		prevProposal *group.Proposal
		curProposal  *group.Proposal
		expBroken    bool
	}{
		"invariant not broken": {
			prevProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        prevCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "1", NoCount: "0", AbstainCount: "0", VetoCount: "0"},
				Timeout:            prevCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},

			curProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr2.String(),
				Proposers:          []string{addr2.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "2", NoCount: "0", AbstainCount: "0", VetoCount: "0"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
		},
		"current block yes vote count must be greater than previous block yes vote count": {
			prevProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        prevCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "2", NoCount: "0", AbstainCount: "0", VetoCount: "0"},
				Timeout:            prevCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			curProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr2.String(),
				Proposers:          []string{addr2.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "1", NoCount: "0", AbstainCount: "0", VetoCount: "0"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			expBroken: true,
		},
		"current block no vote count must be greater than previous block no vote count": {
			prevProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        prevCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "0", NoCount: "2", AbstainCount: "0", VetoCount: "0"},
				Timeout:            prevCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			curProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr2.String(),
				Proposers:          []string{addr2.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "0", NoCount: "1", AbstainCount: "0", VetoCount: "0"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			expBroken: true,
		},
		"current block abstain vote count must be greater than previous block abstain vote count": {
			prevProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        prevCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "0", NoCount: "0", AbstainCount: "2", VetoCount: "0"},
				Timeout:            prevCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			curProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr2.String(),
				Proposers:          []string{addr2.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "0", NoCount: "0", AbstainCount: "1", VetoCount: "0"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			expBroken: true,
		},
		"current block veto vote count must be greater than previous block veto vote count": {
			prevProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        prevCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "0", NoCount: "0", AbstainCount: "0", VetoCount: "2"},
				Timeout:            prevCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			curProposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr2.String(),
				Proposers:          []string{addr2.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "0", NoCount: "0", AbstainCount: "0", VetoCount: "1"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			expBroken: true,
		},
	}

	for _, spec := range specs {

		prevProposal := spec.prevProposal
		curProposal := spec.curProposal

		cachePrevCtx, _ := prevCtx.CacheContext()
		cacheCurCtx, _ := curCtx.CacheContext()

		_, err = proposalTable.Create(cachePrevCtx.KVStore(key), prevProposal)
		s.Require().NoError(err)
		_, err = proposalTable.Create(cacheCurCtx.KVStore(key), curProposal)
		s.Require().NoError(err)

		_, broken := keeper.TallyVotesInvariantHelper(cacheCurCtx, cachePrevCtx, key, *proposalTable)
		s.Require().Equal(spec.expBroken, broken)
	}
}

func (s *invariantTestSuite) TestGroupTotalWeightInvariant() {
	sdkCtx, _ := s.ctx.CacheContext()
	curCtx, cdc, key := sdkCtx, s.cdc, s.key

	// Group Table
	groupTable, err := orm.NewAutoUInt64Table([2]byte{keeper.GroupTablePrefix}, keeper.GroupTableSeqPrefix, &group.GroupInfo{}, cdc)
	s.Require().NoError(err)

	// Group Member Table
	groupMemberTable, err := orm.NewPrimaryKeyTable([2]byte{keeper.GroupMemberTablePrefix}, &group.GroupMember{}, cdc)
	s.Require().NoError(err)

	groupMemberByGroupIndex, err := orm.NewIndex(groupMemberTable, keeper.GroupMemberByGroupIndexPrefix, func(val interface{}) ([]interface{}, error) {
		group := val.(*group.GroupMember).GroupId
		return []interface{}{group}, nil
	}, group.GroupMember{}.GroupId)
	s.Require().NoError(err)

	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	specs := map[string]struct {
		groupsInfo   *group.GroupInfo
		groupMembers []*group.GroupMember
		expBroken    bool
	}{
		"invariant not broken": {
			groupsInfo: &group.GroupInfo{
				GroupId:     1,
				Admin:       addr1.String(),
				Version:     1,
				TotalWeight: "3",
			},
			groupMembers: []*group.GroupMember{
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr1.String(),
						Weight:  "1",
					},
				},
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr2.String(),
						Weight:  "2",
					},
				},
			},
			expBroken: false,
		},

		"group's TotalWeight must be equal to sum of its members weight ": {
			groupsInfo: &group.GroupInfo{
				GroupId:     1,
				Admin:       addr1.String(),
				Version:     1,
				TotalWeight: "3",
			},
			groupMembers: []*group.GroupMember{
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr1.String(),
						Weight:  "2",
					},
				},
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr2.String(),
						Weight:  "2",
					},
				},
			},
			expBroken: true,
		},
	}

	for _, spec := range specs {
		cacheCurCtx, _ := curCtx.CacheContext()
		groupsInfo := spec.groupsInfo
		groupMembers := spec.groupMembers

		_, err := groupTable.Create(cacheCurCtx.KVStore(key), groupsInfo)
		s.Require().NoError(err)

		for i := 0; i < len(groupMembers); i++ {
			err := groupMemberTable.Create(cacheCurCtx.KVStore(key), groupMembers[i])
			s.Require().NoError(err)
		}

		_, broken := keeper.GroupTotalWeightInvariantHelper(cacheCurCtx, key, *groupTable, groupMemberByGroupIndex)
		s.Require().Equal(spec.expBroken, broken)

	}
}

func (s *invariantTestSuite) TestTallyVotesSumInvariant() {
	sdkCtx, _ := s.ctx.CacheContext()
	curCtx, cdc, key := sdkCtx, s.cdc, s.key

	// Group Table
	groupTable, err := orm.NewAutoUInt64Table([2]byte{keeper.GroupTablePrefix}, keeper.GroupTableSeqPrefix, &group.GroupInfo{}, cdc)
	s.Require().NoError(err)

	// Group Policy Table
	groupPolicyTable, err := orm.NewPrimaryKeyTable([2]byte{keeper.GroupPolicyTablePrefix}, &group.GroupPolicyInfo{}, cdc)
	s.Require().NoError(err)

	// Group Member Table
	groupMemberTable, err := orm.NewPrimaryKeyTable([2]byte{keeper.GroupMemberTablePrefix}, &group.GroupMember{}, cdc)
	s.Require().NoError(err)

	// Proposal Table
	proposalTable, err := orm.NewAutoUInt64Table([2]byte{keeper.ProposalTablePrefix}, keeper.ProposalTableSeqPrefix, &group.Proposal{}, cdc)
	s.Require().NoError(err)

	// Vote Table
	voteTable, err := orm.NewPrimaryKeyTable([2]byte{keeper.VoteTablePrefix}, &group.Vote{}, cdc)
	s.Require().NoError(err)

	voteByProposalIndex, err := orm.NewIndex(voteTable, keeper.VoteByProposalIndexPrefix, func(value interface{}) ([]interface{}, error) {
		return []interface{}{value.(*group.Vote).ProposalId}, nil
	}, group.Vote{}.ProposalId)
	s.Require().NoError(err)

	_, _, adminAddr := testdata.KeyTestPubAddr()
	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	specs := map[string]struct {
		groupsInfo   *group.GroupInfo
		groupPolicy  *group.GroupPolicyInfo
		groupMembers []*group.GroupMember
		proposal     *group.Proposal
		votes        []*group.Vote
		expBroken    bool
	}{
		"invariant not broken": {
			groupsInfo: &group.GroupInfo{
				GroupId:     1,
				Admin:       adminAddr.String(),
				Version:     1,
				TotalWeight: "7",
			},
			groupPolicy: &group.GroupPolicyInfo{
				Address: addr1.String(),
				GroupId: 1,
				Admin:   adminAddr.String(),
				Version: 1,
			},
			groupMembers: []*group.GroupMember{
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr1.String(),
						Weight:  "4",
					},
				},
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr2.String(),
						Weight:  "3",
					},
				},
			},
			proposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "4", NoCount: "3", AbstainCount: "0", VetoCount: "0"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			votes: []*group.Vote{
				{
					ProposalId:  1,
					Voter:       addr1.String(),
					Choice:      group.Choice_CHOICE_YES,
					SubmittedAt: curCtx.BlockTime(),
				},
				{
					ProposalId:  1,
					Voter:       addr2.String(),
					Choice:      group.Choice_CHOICE_NO,
					SubmittedAt: curCtx.BlockTime(),
				},
			},
			expBroken: false,
		},
		"proposal tally must correspond to the sum of vote weights": {
			groupsInfo: &group.GroupInfo{
				GroupId:     1,
				Admin:       adminAddr.String(),
				Version:     1,
				TotalWeight: "5",
			},
			groupPolicy: &group.GroupPolicyInfo{
				Address: addr1.String(),
				GroupId: 1,
				Admin:   adminAddr.String(),
				Version: 1,
			},
			groupMembers: []*group.GroupMember{
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr1.String(),
						Weight:  "2",
					},
				},
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr2.String(),
						Weight:  "3",
					},
				},
			},
			proposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "6", NoCount: "0", AbstainCount: "0", VetoCount: "0"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			votes: []*group.Vote{
				{
					ProposalId:  1,
					Voter:       addr1.String(),
					Choice:      group.Choice_CHOICE_YES,
					SubmittedAt: curCtx.BlockTime(),
				},
				{
					ProposalId:  1,
					Voter:       addr2.String(),
					Choice:      group.Choice_CHOICE_YES,
					SubmittedAt: curCtx.BlockTime(),
				},
			},
			expBroken: true,
		},
		"proposal VoteState must correspond to the vote choice": {
			groupsInfo: &group.GroupInfo{
				GroupId:     1,
				Admin:       adminAddr.String(),
				Version:     1,
				TotalWeight: "7",
			},
			groupPolicy: &group.GroupPolicyInfo{
				Address: addr1.String(),
				GroupId: 1,
				Admin:   adminAddr.String(),
				Version: 1,
			},
			groupMembers: []*group.GroupMember{
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr1.String(),
						Weight:  "4",
					},
				},
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr2.String(),
						Weight:  "3",
					},
				},
			},
			proposal: &group.Proposal{
				ProposalId:         1,
				Address:            addr1.String(),
				Proposers:          []string{addr1.String()},
				SubmittedAt:        curCtx.BlockTime(),
				GroupVersion:       1,
				GroupPolicyVersion: 1,
				Status:             group.ProposalStatusSubmitted,
				Result:             group.ProposalResultUnfinalized,
				VoteState:          group.Tally{YesCount: "4", NoCount: "3", AbstainCount: "0", VetoCount: "0"},
				Timeout:            curCtx.BlockTime().Add(time.Second * 600),
				ExecutorResult:     group.ProposalExecutorResultNotRun,
			},
			votes: []*group.Vote{
				{
					ProposalId:  1,
					Voter:       addr1.String(),
					Choice:      group.Choice_CHOICE_YES,
					SubmittedAt: curCtx.BlockTime(),
				},
				{
					ProposalId:  1,
					Voter:       addr2.String(),
					Choice:      group.Choice_CHOICE_ABSTAIN,
					SubmittedAt: curCtx.BlockTime(),
				},
			},
			expBroken: true,
		},
	}

	for _, spec := range specs {
		cacheCurCtx, _ := curCtx.CacheContext()
		groupsInfo := spec.groupsInfo
		proposal := spec.proposal
		groupPolicy := spec.groupPolicy
		groupMembers := spec.groupMembers
		votes := spec.votes

		_, err := groupTable.Create(cacheCurCtx.KVStore(key), groupsInfo)
		s.Require().NoError(err)

		err = groupPolicy.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Second))
		s.Require().NoError(err)
		err = groupPolicyTable.Create(cacheCurCtx.KVStore(key), groupPolicy)
		s.Require().NoError(err)

		for i := 0; i < len(groupMembers); i++ {
			err = groupMemberTable.Create(cacheCurCtx.KVStore(key), groupMembers[i])
			s.Require().NoError(err)
		}

		_, err = proposalTable.Create(cacheCurCtx.KVStore(key), proposal)
		s.Require().NoError(err)

		for i := 0; i < len(votes); i++ {
			err = voteTable.Create(cacheCurCtx.KVStore(key), votes[i])
			s.Require().NoError(err)
		}

		_, broken := keeper.TallyVotesSumInvariantHelper(cacheCurCtx, key, *groupTable, *proposalTable, *groupMemberTable, voteByProposalIndex, *groupPolicyTable)
		s.Require().Equal(spec.expBroken, broken)
	}
}
