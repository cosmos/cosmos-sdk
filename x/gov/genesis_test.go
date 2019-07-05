package gov

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestEqualProposalID(t *testing.T) {
	state1 := GenesisState{}
	state2 := GenesisState{}
	require.Equal(t, state1, state2)

	// Proposals
	state1.StartingProposalID = 1
	require.NotEqual(t, state1, state2)
	require.False(t, state1.Equal(state2))

	state2.StartingProposalID = 1
	require.Equal(t, state1, state2)
	require.True(t, state1.Equal(state2))
}

func TestEqualProposals(t *testing.T) {
	// Generate mock app and keepers
	input := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(input.addrs)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	// Submit two proposals
	proposal := testProposal()
	proposal1, err := input.keeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)
	proposal2, err := input.keeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)

	// They are similar but their IDs should be different
	require.NotEqual(t, proposal1, proposal2)
	require.False(t, ProposalEqual(proposal1, proposal2))

	// Now create two genesis blocks
	state1 := GenesisState{Proposals: []Proposal{proposal1}}
	state2 := GenesisState{Proposals: []Proposal{proposal2}}
	require.NotEqual(t, state1, state2)
	require.False(t, state1.Equal(state2))

	// Now make proposals identical by setting both IDs to 55
	proposal1.ProposalID = 55
	proposal2.ProposalID = 55
	require.Equal(t, proposal1, proposal1)
	require.True(t, ProposalEqual(proposal1, proposal2))

	// Reassign proposals into state
	state1.Proposals[0] = proposal1
	state2.Proposals[0] = proposal2

	// State should be identical now..
	require.Equal(t, state1, state2)
	require.True(t, state1.Equal(state2))
}

func TestImportExportQueues(t *testing.T) {
	// Generate mock app and keepers
	input := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(input.addrs)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	// Create two proposals, put the second into the voting period
	proposal := testProposal()
	proposal1, err := input.keeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)
	proposalID1 := proposal1.ProposalID

	proposal2, err := input.keeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)
	proposalID2 := proposal2.ProposalID

	err, votingStarted := input.keeper.AddDeposit(ctx, proposalID2, input.addrs[0], input.keeper.GetDepositParams(ctx).MinDeposit)
	require.NoError(t, err)
	require.True(t, votingStarted)

	proposal1, ok := input.keeper.GetProposal(ctx, proposalID1)
	require.True(t, ok)
	proposal2, ok = input.keeper.GetProposal(ctx, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == StatusDepositPeriod)
	require.True(t, proposal2.Status == StatusVotingPeriod)

	genAccs := input.mApp.AccountKeeper.GetAllAccounts(ctx)

	// Export the state and import it into a new Mock App
	genState := ExportGenesis(ctx, input.keeper)
	input2 := getMockApp(t, 2, genState, genAccs)

	header = abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input2.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx2 := input2.mApp.BaseApp.NewContext(false, abci.Header{})

	// Jump the time forward past the DepositPeriod and VotingPeriod
	ctx2 = ctx2.WithBlockTime(ctx2.BlockHeader().Time.Add(input2.keeper.GetDepositParams(ctx2).MaxDepositPeriod).Add(input2.keeper.GetVotingParams(ctx2).VotingPeriod))

	// Make sure that they are still in the DepositPeriod and VotingPeriod respectively
	proposal1, ok = input2.keeper.GetProposal(ctx2, proposalID1)
	require.True(t, ok)
	proposal2, ok = input2.keeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == StatusDepositPeriod)
	require.True(t, proposal2.Status == StatusVotingPeriod)

	require.Equal(t, input2.keeper.GetDepositParams(ctx2).MinDeposit, input2.keeper.GetGovernanceAccount(ctx2).GetCoins())

	// Run the endblocker. Check to make sure that proposal1 is removed from state, and proposal2 is finished VotingPeriod.
	EndBlocker(ctx2, input2.keeper)

	proposal1, ok = input2.keeper.GetProposal(ctx2, proposalID1)
	require.False(t, ok)
	proposal2, ok = input2.keeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal2.Status == StatusRejected)
}
