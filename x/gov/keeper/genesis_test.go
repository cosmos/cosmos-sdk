package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (suite *KeeperTestSuite) TestImportExportQueues_ErrorInconsistentState() {
	suite.reset()
	suite.acctKeeper.EXPECT().SetModuleAccount(suite.ctx, gomock.Any()).AnyTimes()
	suite.Require().Panics(func() {
		keeper.InitGenesis(suite.ctx, suite.acctKeeper, suite.bankKeeper, suite.govKeeper, &v1.GenesisState{
			Deposits: v1.Deposits{
				{
					ProposalId: 1234,
					Depositor:  "me",
					Amount: sdk.Coins{
						sdk.NewCoin(
							"stake",
							sdkmath.NewInt(1234),
						),
					},
				},
			},
		})
	})
	keeper.InitGenesis(suite.ctx, suite.acctKeeper, suite.bankKeeper, suite.govKeeper, v1.DefaultGenesisState())
	genState, err := keeper.ExportGenesis(suite.ctx, suite.govKeeper)
	suite.Require().NoError(err)
	suite.Require().Equal(genState, v1.DefaultGenesisState())
}

// TestInitGenesis_PanicsWhenDistrCancelDestSetWithoutDistrKeeper covers the
// guard added in #25616 that requires a non-nil distribution keeper whenever
// the distribution module address is configured as the proposal cancel
// destination.
func TestInitGenesis_PanicsWhenDistrCancelDestSetWithoutDistrKeeper(t *testing.T) {
	_, acctKeeper, bankKeeper, stakingKeeper, _, storeService, encCfg, ctx := setupGovKeeper(t)

	// Build a separate keeper sharing the same dependencies but with a nil
	// distribution keeper, so we can hit the guard without disturbing the
	// suite-wide keeper.
	govKeeperNoDistr := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		acctKeeper,
		bankKeeper,
		nil,
		baseapp.NewMsgServiceRouter(),
		types.DefaultConfig(),
		govAcct.String(),
		keeper.NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper),
	)

	params := v1.DefaultParams()
	params.ProposalCancelDest = distAcct.String()

	expected := fmt.Sprintf(
		"must set DistrKeeper first if using distribution module (%s) as proposal cancel destination",
		distAcct.String(),
	)
	require.PanicsWithValue(t, expected, func() {
		keeper.InitGenesis(ctx, acctKeeper, bankKeeper, govKeeperNoDistr, &v1.GenesisState{
			StartingProposalId: 1,
			Params:             &params,
		})
	})
}

// TestInitGenesis_RoundTripProposalsAndVotes verifies that proposals in both
// deposit and voting period statuses, plus a vote, survive an InitGenesis →
// ExportGenesis round-trip.
func (suite *KeeperTestSuite) TestInitGenesis_RoundTripProposalsAndVotes() {
	suite.reset()
	suite.acctKeeper.EXPECT().SetModuleAccount(suite.ctx, gomock.Any()).AnyTimes()

	blockTime := suite.ctx.BlockHeader().Time
	depositEnd := blockTime.Add(time.Hour)
	votingEnd := blockTime.Add(2 * time.Hour)

	proposalA := &v1.Proposal{
		Id:             1,
		Status:         v1.StatusDepositPeriod,
		SubmitTime:     &blockTime,
		DepositEndTime: &depositEnd,
		Title:          "proposal A",
		Summary:        "summary A",
		Proposer:       suite.addrs[0].String(),
	}
	proposalB := &v1.Proposal{
		Id:              2,
		Status:          v1.StatusVotingPeriod,
		SubmitTime:      &blockTime,
		DepositEndTime:  &depositEnd,
		VotingStartTime: &blockTime,
		VotingEndTime:   &votingEnd,
		Title:           "proposal B",
		Summary:         "summary B",
		Proposer:        suite.addrs[1].String(),
	}
	vote := &v1.Vote{
		ProposalId: 2,
		Voter:      suite.addrs[0].String(),
		Options:    v1.NewNonSplitVoteOption(v1.OptionYes),
	}

	params := v1.DefaultParams()
	state := &v1.GenesisState{
		StartingProposalId: 3,
		Params:             &params,
		Constitution:       "test constitution",
		Proposals:          v1.Proposals{proposalA, proposalB},
		Votes:              v1.Votes{vote},
	}

	suite.Require().NotPanics(func() {
		keeper.InitGenesis(suite.ctx, suite.acctKeeper, suite.bankKeeper, suite.govKeeper, state)
	})

	exported, err := keeper.ExportGenesis(suite.ctx, suite.govKeeper)
	suite.Require().NoError(err)

	suite.Require().Equal("test constitution", exported.Constitution)
	suite.Require().Equal(uint64(3), exported.StartingProposalId)

	// Both proposals should be exported, preserving their IDs and statuses.
	suite.Require().Len(exported.Proposals, 2)
	exportedByID := make(map[uint64]*v1.Proposal, len(exported.Proposals))
	for _, p := range exported.Proposals {
		exportedByID[p.Id] = p
	}
	suite.Require().Contains(exportedByID, uint64(1))
	suite.Require().Equal(v1.StatusDepositPeriod, exportedByID[1].Status)
	suite.Require().Contains(exportedByID, uint64(2))
	suite.Require().Equal(v1.StatusVotingPeriod, exportedByID[2].Status)

	// The single vote should round-trip.
	suite.Require().Len(exported.Votes, 1)
	suite.Require().Equal(uint64(2), exported.Votes[0].ProposalId)
	suite.Require().Equal(suite.addrs[0].String(), exported.Votes[0].Voter)
}
