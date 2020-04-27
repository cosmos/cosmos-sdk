package test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/cli_test/helpers"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankCli "github.com/cosmos/cosmos-sdk/x/bank/client/cli_test"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimCLISubmitProposal(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	defer proc.Stop(false)

	QueryGovParamDeposit(f)
	QueryGovParamVoting(f)
	QueryGovParamTallying(f)

	fooAddr := f.KeyAddress(helpers.KeyFoo)

	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, bankCli.QueryBalances(f, fooAddr).AmountOf(sdk.DefaultBondDenom))

	proposalsQuery := QueryGovProposals(f)
	require.Empty(t, proposalsQuery)

	// Test submit generate only for submit proposal
	proposalTokens := sdk.TokensFromConsensusPower(5)
	success, stdout, stderr := TxGovSubmitProposal(f,
		fooAddr.String(), "Text", "Test", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "--generate-only", "-y")
	require.True(t, success)
	require.Empty(t, stderr)
	msg := helpers.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = TxGovSubmitProposal(f, helpers.KeyFoo, "Text", "Test", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "--dry-run")
	require.True(t, success)

	// Create the proposal
	TxGovSubmitProposal(f, helpers.KeyFoo, "Text", "Test", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure transaction events can be queried
	searchResult := f.QueryTxs(1, 50, "message.action=submit_proposal", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure deposit was deducted
	require.Equal(t, startTokens.Sub(proposalTokens), bankCli.QueryBalances(f, fooAddr).AmountOf(helpers.Denom))

	// Ensure propsal is directly queryable
	proposal1 := QueryGovProposal(f, 1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusDepositPeriod, proposal1.Status)

	// Ensure query proposals returns properly
	proposalsQuery = QueryGovProposals(f)
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// Query the deposits on the proposal
	deposit := QueryGovDeposit(f, 1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(helpers.Denom))

	// Test deposit generate only
	depositTokens := sdk.TokensFromConsensusPower(10)
	success, stdout, stderr = TxGovDeposit(f, 1, fooAddr.String(), sdk.NewCoin(helpers.Denom, depositTokens), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = helpers.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Run the deposit transaction
	TxGovDeposit(f, 1, helpers.KeyFoo, sdk.NewCoin(helpers.Denom, depositTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// test query deposit
	deposits := QueryGovDeposits(f, 1)
	require.Len(t, deposits, 1)
	require.Equal(t, proposalTokens.Add(depositTokens), deposits[0].Amount.AmountOf(helpers.Denom))

	// Ensure querying the deposit returns the proper amount
	deposit = QueryGovDeposit(f, 1, fooAddr)
	require.Equal(t, proposalTokens.Add(depositTokens), deposit.Amount.AmountOf(helpers.Denom))

	// Ensure events are set on the transaction
	searchResult = f.QueryTxs(1, 50, "message.action=deposit", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure account has expected amount of funds
	require.Equal(t, startTokens.Sub(proposalTokens.Add(depositTokens)), bankCli.QueryBalances(f, fooAddr).AmountOf(helpers.Denom))

	// Fetch the proposal and ensure it is now in the voting period
	proposal1 = QueryGovProposal(f, 1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusVotingPeriod, proposal1.Status)

	// Test vote generate only
	success, stdout, stderr = TxGovVote(f, 1, gov.OptionYes, fooAddr.String(), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = helpers.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Vote on the proposal
	TxGovVote(f, 1, gov.OptionYes, helpers.KeyFoo, "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Query the vote
	vote := QueryGovVote(f, 1, fooAddr)
	require.Equal(t, uint64(1), vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	// Query the votes
	votes := QueryGovVotes(f, 1)
	require.Len(t, votes, 1)
	require.Equal(t, uint64(1), votes[0].ProposalID)
	require.Equal(t, gov.OptionYes, votes[0].Option)

	// Ensure events are applied to voting transaction properly
	searchResult = f.QueryTxs(1, 50, "message.action=vote", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure no proposals in deposit period
	proposalsQuery = QueryGovProposals(f, "--status=DepositPeriod")
	require.Empty(t, proposalsQuery)

	// Ensure the proposal returns as in the voting period
	proposalsQuery = QueryGovProposals(f, "--status=VotingPeriod")
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// submit a second test proposal
	TxGovSubmitProposal(f, helpers.KeyFoo, "Text", "Apples", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Test limit on proposals query
	proposalsQuery = QueryGovProposals(f, "--limit=2")
	require.Len(t, proposalsQuery, 2)
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	f.Cleanup()
}
