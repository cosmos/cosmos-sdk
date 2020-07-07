// +build cli_test

package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutils "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestCLISubmitProposal(t *testing.T) {
	t.SkipNow() // TODO: Bring back once viper is refactored.
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	testutil.QueryGovParamDeposit(f)
	testutil.QueryGovParamVoting(f)
	testutil.QueryGovParamTallying(f)

	fooAddr := f.KeyAddress(cli.KeyFoo)

	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, banktestutils.QueryBalances(f, fooAddr).AmountOf(sdk.DefaultBondDenom))

	proposalsQuery := testutil.QueryGovProposals(f)
	require.Empty(t, proposalsQuery)

	// Test submit generate only for submit proposal
	proposalTokens := sdk.TokensFromConsensusPower(5)
	success, stdout, stderr := testutil.TxGovSubmitProposal(f,
		fooAddr.String(), "Text", "Test", "test", sdk.NewCoin(cli.Denom, proposalTokens), "--generate-only", "-y")
	require.True(t, success)
	require.Empty(t, stderr)
	msg := cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = testutil.TxGovSubmitProposal(f, cli.KeyFoo, "Text", "Test", "test", sdk.NewCoin(cli.Denom, proposalTokens), "--dry-run")
	require.True(t, success)

	// Create the proposal
	testutil.TxGovSubmitProposal(f, cli.KeyFoo, "Text", "Test", "test", sdk.NewCoin(cli.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure transaction events can be queried
	searchResult := f.QueryTxs(1, 50, "message.action=submit_proposal", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure deposit was deducted
	require.Equal(t, startTokens.Sub(proposalTokens), banktestutils.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	// Ensure propsal is directly queryable
	proposal1 := testutil.QueryGovProposal(f, 1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, types.StatusDepositPeriod, proposal1.Status)

	// Ensure query proposals returns properly
	proposalsQuery = testutil.QueryGovProposals(f)
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// Query the deposits on the proposal
	deposit := testutil.QueryGovDeposit(f, 1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(cli.Denom))

	// Test deposit generate only
	depositTokens := sdk.TokensFromConsensusPower(10)
	success, stdout, stderr = testutil.TxGovDeposit(f, 1, fooAddr.String(), sdk.NewCoin(cli.Denom, depositTokens), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Run the deposit transaction
	testutil.TxGovDeposit(f, 1, cli.KeyFoo, sdk.NewCoin(cli.Denom, depositTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// test query deposit
	deposits := testutil.QueryGovDeposits(f, 1)
	require.Len(t, deposits, 1)
	require.Equal(t, proposalTokens.Add(depositTokens), deposits[0].Amount.AmountOf(cli.Denom))

	// Ensure querying the deposit returns the proper amount
	deposit = testutil.QueryGovDeposit(f, 1, fooAddr)
	require.Equal(t, proposalTokens.Add(depositTokens), deposit.Amount.AmountOf(cli.Denom))

	// Ensure events are set on the transaction
	searchResult = f.QueryTxs(1, 50, "message.action=deposit", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure account has expected amount of funds
	require.Equal(t, startTokens.Sub(proposalTokens.Add(depositTokens)), banktestutils.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	// Fetch the proposal and ensure it is now in the voting period
	proposal1 = testutil.QueryGovProposal(f, 1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, types.StatusVotingPeriod, proposal1.Status)

	// Test vote generate only
	success, stdout, stderr = testutil.TxGovVote(f, 1, types.OptionYes, fooAddr.String(), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Vote on the proposal
	testutil.TxGovVote(f, 1, types.OptionYes, cli.KeyFoo, "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Query the vote
	vote := testutil.QueryGovVote(f, 1, fooAddr)
	require.Equal(t, uint64(1), vote.ProposalID)
	require.Equal(t, types.OptionYes, vote.Option)

	// Query the votes
	votes := testutil.QueryGovVotes(f, 1)
	require.Len(t, votes, 1)
	require.Equal(t, uint64(1), votes[0].ProposalID)
	require.Equal(t, types.OptionYes, votes[0].Option)

	// Ensure events are applied to voting transaction properly
	searchResult = f.QueryTxs(1, 50, "message.action=vote", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure no proposals in deposit period
	proposalsQuery = testutil.QueryGovProposals(f, "--status=DepositPeriod")
	require.Empty(t, proposalsQuery)

	// Ensure the proposal returns as in the voting period
	proposalsQuery = testutil.QueryGovProposals(f, "--status=VotingPeriod")
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// submit a second test proposal
	testutil.TxGovSubmitProposal(f, cli.KeyFoo, "Text", "Apples", "test", sdk.NewCoin(cli.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Test limit on proposals query
	proposalsQuery = testutil.QueryGovProposals(f, "--limit=2")
	require.Len(t, proposalsQuery, 2)
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	f.Cleanup()
}
