package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetProposal(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := keeper.submitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	keeper.SetProposal(ctx, proposal)

	gotProposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, ProposalEqual(proposal, gotProposal))
}

func TestIncrementProposalNumber(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	keeper.submitProposal(ctx, tp)
	keeper.submitProposal(ctx, tp)
	keeper.submitProposal(ctx, tp)
	keeper.submitProposal(ctx, tp)
	keeper.submitProposal(ctx, tp)
	proposal6, err := keeper.submitProposal(ctx, tp)
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.ProposalID)
}

func TestActivateVotingPeriod(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := keeper.submitProposal(ctx, tp)
	require.NoError(t, err)

	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	keeper.activateVotingPeriod(ctx, proposal)

	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	proposal, ok := keeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := keeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())
	var proposalID uint64
	keeper.cdc.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.ProposalID)
	activeIterator.Close()
}

func TestDeposits(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(addrs)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := keeper.submitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	fourSteak := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(4)))
	fiveSteak := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(5)))

	addr0Initial := keeper.ck.GetCoins(ctx, addrs[0])
	addr1Initial := keeper.ck.GetCoins(ctx, addrs[1])

	expTokens := sdk.TokensFromTendermintPower(42)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, expTokens)), addr0Initial)
	require.True(t, proposal.TotalDeposit.IsEqual(sdk.NewCoins()))

	// Check no deposits at beginning
	deposit, found := keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.False(t, found)
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	// Check first deposit
	err, votingStarted := keeper.AddDeposit(ctx, proposalID, addrs[0], fourSteak)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourSteak, deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositor)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourSteak, proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check a second deposit from same address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveSteak)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourSteak.Add(fiveSteak), deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositor)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourSteak.Add(fiveSteak), proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourSteak).Sub(fiveSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check third deposit from a new address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[1], fourSteak)
	require.Nil(t, err)
	require.True(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1], deposit.Depositor)
	require.Equal(t, fourSteak, deposit.Amount)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourSteak.Add(fiveSteak).Add(fourSteak), proposal.TotalDeposit)
	require.Equal(t, addr1Initial.Sub(fourSteak), keeper.ck.GetCoins(ctx, addrs[1]))

	// Check that proposal moved to voting period
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	// Test deposit iterator
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, addrs[0], deposit.Depositor)
	require.Equal(t, fourSteak.Add(fiveSteak), deposit.Amount)
	depositsIterator.Next()
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, addrs[1], deposit.Depositor)
	require.Equal(t, fourSteak, deposit.Amount)
	depositsIterator.Next()
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()

	// Test Refund Deposits
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, fourSteak, deposit.Amount)
	keeper.RefundDeposits(ctx, proposalID)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.False(t, found)
	require.Equal(t, addr0Initial, keeper.ck.GetCoins(ctx, addrs[0]))
	require.Equal(t, addr1Initial, keeper.ck.GetCoins(ctx, addrs[1]))

}

func TestVotes(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(addrs)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := keeper.submitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// Test first vote
	keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	vote, found := keeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionAbstain, vote.Option)

	// Test change of vote
	keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	vote, found = keeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionYes, vote.Option)

	// Test second vote
	keeper.AddVote(ctx, proposalID, addrs[1], OptionNoWithVeto)
	vote, found = keeper.GetVote(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionNoWithVeto, vote.Option)

	// Test vote iterator
	votesIterator := keeper.GetVotes(ctx, proposalID)
	require.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionYes, vote.Option)
	votesIterator.Next()
	require.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, addrs[1], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionNoWithVeto, vote.Option)
	votesIterator.Next()
	require.False(t, votesIterator.Valid())
	votesIterator.Close()
}

func TestProposalQueues(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	mapp.InitChainer(ctx, abci.RequestInitChain{})

	// create test proposals
	tp := testProposal()
	proposal, err := keeper.submitProposal(ctx, tp)
	require.NoError(t, err)

	inactiveIterator := keeper.InactiveProposalQueueIterator(ctx, proposal.DepositEndTime)
	require.True(t, inactiveIterator.Valid())
	var proposalID uint64
	keeper.cdc.UnmarshalBinaryLengthPrefixed(inactiveIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.ProposalID)
	inactiveIterator.Close()

	keeper.activateVotingPeriod(ctx, proposal)

	proposal, ok := keeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := keeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())
	keeper.cdc.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.ProposalID)
	activeIterator.Close()
}
