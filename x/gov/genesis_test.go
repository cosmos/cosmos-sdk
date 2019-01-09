package gov

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestImportExportQueues(t *testing.T) {

	// Generate mock app and keepers
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(addrs)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	// Create two proposals, put the second into the voting period
	proposal1 := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID1 := proposal1.GetProposalID()

	proposal2 := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID2 := proposal2.GetProposalID()

	_, votingStarted := keeper.AddDeposit(ctx, proposalID2, addrs[0], keeper.GetDepositParams(ctx).MinDeposit)
	require.True(t, votingStarted)

	require.True(t, keeper.GetProposal(ctx, proposalID1).GetStatus() == StatusDepositPeriod)
	require.True(t, keeper.GetProposal(ctx, proposalID2).GetStatus() == StatusVotingPeriod)

	genAccs := mock.GetAllAccounts(mapp.AccountKeeper, ctx)

	// Export the state and import it into a new Mock App
	genState := ExportGenesis(ctx, keeper)
	mapp2, keeper2, _, _, _, _ := getMockApp(t, 2, genState, genAccs)

	mapp2.BeginBlock(abci.RequestBeginBlock{})
	ctx2 := mapp2.BaseApp.NewContext(false, abci.Header{})

	// Jump the time forward past the DepositPeriod and VotingPeriod
	ctx2 = ctx2.WithBlockTime(ctx2.BlockHeader().Time.Add(keeper2.GetDepositParams(ctx2).MaxDepositPeriod).Add(keeper2.GetVotingParams(ctx2).VotingPeriod))

	// Make sure that they are still in the DepositPeriod and VotingPeriod respectively
	require.True(t, keeper2.GetProposal(ctx2, proposalID1).GetStatus() == StatusDepositPeriod)
	require.True(t, keeper2.GetProposal(ctx2, proposalID2).GetStatus() == StatusVotingPeriod)

	// Run the endblocker.  Check to make sure that proposal1 is removed from state, and proposal2 is finished VotingPeriod.
	EndBlocker(ctx2, keeper2)

	require.Nil(t, keeper2.GetProposal(ctx2, proposalID1))
	require.True(t, keeper2.GetProposal(ctx2, proposalID2).GetStatus() == StatusRejected)
}
