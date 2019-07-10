package gov

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetProposal(t *testing.T) {
	input := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	input.keeper.SetProposal(ctx, proposal)

	gotProposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, ProposalEqual(proposal, gotProposal))
}

func TestIncrementProposalNumber(t *testing.T) {
	input := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	input.keeper.SubmitProposal(ctx, tp)
	input.keeper.SubmitProposal(ctx, tp)
	input.keeper.SubmitProposal(ctx, tp)
	input.keeper.SubmitProposal(ctx, tp)
	input.keeper.SubmitProposal(ctx, tp)
	proposal6, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.ProposalID)
}

func TestActivateVotingPeriod(t *testing.T) {
	input := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	input.keeper.activateVotingPeriod(ctx, proposal)

	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	proposal, ok := input.keeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := input.keeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())
	var proposalID uint64
	input.keeper.cdc.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.ProposalID)
	activeIterator.Close()
}

func TestDeposits(t *testing.T) {
	input := getMockApp(t, 2, GenesisState{}, nil)

	SortAddresses(input.addrs)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	fourStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(4)))
	fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5)))

	addr0Initial := input.mApp.AccountKeeper.GetAccount(ctx, input.addrs[0]).GetCoins()
	addr1Initial := input.mApp.AccountKeeper.GetAccount(ctx, input.addrs[1]).GetCoins()

	expTokens := sdk.TokensFromConsensusPower(42)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, expTokens)), addr0Initial)
	require.True(t, proposal.TotalDeposit.IsEqual(sdk.NewCoins()))

	// Check no deposits at beginning
	deposit, found := input.keeper.GetDeposit(ctx, proposalID, input.addrs[1])
	require.False(t, found)
	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	// Check first deposit
	err, votingStarted := input.keeper.AddDeposit(ctx, proposalID, input.addrs[0], fourStake)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = input.keeper.GetDeposit(ctx, proposalID, input.addrs[0])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	require.Equal(t, input.addrs[0], deposit.Depositor)
	proposal, ok = input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake, proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourStake), input.mApp.AccountKeeper.GetAccount(ctx, input.addrs[0]).GetCoins())

	// Check a second deposit from same address
	err, votingStarted = input.keeper.AddDeposit(ctx, proposalID, input.addrs[0], fiveStake)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = input.keeper.GetDeposit(ctx, proposalID, input.addrs[0])
	require.True(t, found)
	require.Equal(t, fourStake.Add(fiveStake), deposit.Amount)
	require.Equal(t, input.addrs[0], deposit.Depositor)
	proposal, ok = input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake), proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourStake).Sub(fiveStake), input.mApp.AccountKeeper.GetAccount(ctx, input.addrs[0]).GetCoins())

	// Check third deposit from a new address
	err, votingStarted = input.keeper.AddDeposit(ctx, proposalID, input.addrs[1], fourStake)
	require.Nil(t, err)
	require.True(t, votingStarted)
	deposit, found = input.keeper.GetDeposit(ctx, proposalID, input.addrs[1])
	require.True(t, found)
	require.Equal(t, input.addrs[1], deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	proposal, ok = input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake).Add(fourStake), proposal.TotalDeposit)
	require.Equal(t, addr1Initial.Sub(fourStake), input.mApp.AccountKeeper.GetAccount(ctx, input.addrs[1]).GetCoins())

	// Check that proposal moved to voting period
	proposal, ok = input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	// Test deposit iterator
	depositsIterator := input.keeper.GetDepositsIterator(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	input.keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, input.addrs[0], deposit.Depositor)
	require.Equal(t, fourStake.Add(fiveStake), deposit.Amount)
	depositsIterator.Next()
	input.keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, input.addrs[1], deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	depositsIterator.Next()
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()

	// Test Refund Deposits
	deposit, found = input.keeper.GetDeposit(ctx, proposalID, input.addrs[1])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	input.keeper.RefundDeposits(ctx, proposalID)
	deposit, found = input.keeper.GetDeposit(ctx, proposalID, input.addrs[1])
	require.False(t, found)
	require.Equal(t, addr0Initial, input.mApp.AccountKeeper.GetAccount(ctx, input.addrs[0]).GetCoins())
	require.Equal(t, addr1Initial, input.mApp.AccountKeeper.GetAccount(ctx, input.addrs[1]).GetCoins())

}

func TestVotes(t *testing.T) {
	input := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(input.addrs)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	// Test first vote
	input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionAbstain)
	vote, found := input.keeper.GetVote(ctx, proposalID, input.addrs[0])
	require.True(t, found)
	require.Equal(t, input.addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionAbstain, vote.Option)

	// Test change of vote
	input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	vote, found = input.keeper.GetVote(ctx, proposalID, input.addrs[0])
	require.True(t, found)
	require.Equal(t, input.addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionYes, vote.Option)

	// Test second vote
	input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionNoWithVeto)
	vote, found = input.keeper.GetVote(ctx, proposalID, input.addrs[1])
	require.True(t, found)
	require.Equal(t, input.addrs[1], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionNoWithVeto, vote.Option)

	// Test vote iterator
	votesIterator := input.keeper.GetVotesIterator(ctx, proposalID)
	require.True(t, votesIterator.Valid())
	input.keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, input.addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionYes, vote.Option)
	votesIterator.Next()
	require.True(t, votesIterator.Valid())
	input.keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, input.addrs[1], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionNoWithVeto, vote.Option)
	votesIterator.Next()
	require.False(t, votesIterator.Valid())
	votesIterator.Close()
}

func TestProposalQueues(t *testing.T) {
	input := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	input.mApp.InitChainer(ctx, abci.RequestInitChain{})

	// create test proposals
	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	inactiveIterator := input.keeper.InactiveProposalQueueIterator(ctx, proposal.DepositEndTime)
	require.True(t, inactiveIterator.Valid())
	var proposalID uint64
	input.keeper.cdc.UnmarshalBinaryLengthPrefixed(inactiveIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.ProposalID)
	inactiveIterator.Close()

	input.keeper.activateVotingPeriod(ctx, proposal)

	proposal, ok := input.keeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := input.keeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())
	input.keeper.cdc.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.ProposalID)
	activeIterator.Close()
}

type validProposal struct{}

func (validProposal) GetTitle() string         { return "title" }
func (validProposal) GetDescription() string   { return "description" }
func (validProposal) ProposalRoute() string    { return RouterKey }
func (validProposal) ProposalType() string     { return ProposalTypeText }
func (validProposal) String() string           { return "" }
func (validProposal) ValidateBasic() sdk.Error { return nil }

type invalidProposalTitle1 struct{ validProposal }

func (invalidProposalTitle1) GetTitle() string { return "" }

type invalidProposalTitle2 struct{ validProposal }

func (invalidProposalTitle2) GetTitle() string { return strings.Repeat("1234567890", 100) }

type invalidProposalDesc1 struct{ validProposal }

func (invalidProposalDesc1) GetDescription() string { return "" }

type invalidProposalDesc2 struct{ validProposal }

func (invalidProposalDesc2) GetDescription() string { return strings.Repeat("1234567890", 1000) }

type invalidProposalRoute struct{ validProposal }

func (invalidProposalRoute) ProposalRoute() string { return "nonexistingroute" }

type invalidProposalValidation struct{ validProposal }

func (invalidProposalValidation) ValidateBasic() sdk.Error {
	return sdk.NewError(sdk.CodespaceUndefined, sdk.CodeInternal, "")
}

func registerTestCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(validProposal{}, "test/validproposal", nil)
	cdc.RegisterConcrete(invalidProposalTitle1{}, "test/invalidproposalt1", nil)
	cdc.RegisterConcrete(invalidProposalTitle2{}, "test/invalidproposalt2", nil)
	cdc.RegisterConcrete(invalidProposalDesc1{}, "test/invalidproposald1", nil)
	cdc.RegisterConcrete(invalidProposalDesc2{}, "test/invalidproposald2", nil)
	cdc.RegisterConcrete(invalidProposalRoute{}, "test/invalidproposalr", nil)
	cdc.RegisterConcrete(invalidProposalValidation{}, "test/invalidproposalv", nil)
}

func TestSubmitProposal(t *testing.T) {
	input := getMockApp(t, 0, GenesisState{}, nil)

	registerTestCodec(input.keeper.cdc)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	input.mApp.InitChainer(ctx, abci.RequestInitChain{})

	testCases := []struct {
		content     Content
		expectedErr sdk.Error
	}{
		{validProposal{}, nil},
		// Keeper does not check the validity of title and description, no error
		{invalidProposalTitle1{}, nil},
		{invalidProposalTitle2{}, nil},
		{invalidProposalDesc1{}, nil},
		{invalidProposalDesc2{}, nil},
		// error only when invalid route
		{invalidProposalRoute{}, ErrNoProposalHandlerExists(DefaultCodespace, invalidProposalRoute{})},
		// Keeper does not call ValidateBasic, msg.ValidateBasic does
		{invalidProposalValidation{}, nil},
	}

	for _, tc := range testCases {
		_, err := input.keeper.SubmitProposal(ctx, tc.content)
		require.Equal(t, tc.expectedErr, err, "unexpected type of error: %s", err)
	}
}
