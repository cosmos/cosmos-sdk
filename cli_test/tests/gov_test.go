package tests

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
	"path/filepath"
	"testing"
	"github.com/cosmos/cosmos-sdk/cli_test/helpers"

)

func TestGaiaCLISubmitProposal(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start gaiad server
	proc := f.SDStart()
	defer proc.Stop(false)

	f.QueryGovParamDeposit()
	f.QueryGovParamVoting()
	f.QueryGovParamTallying()

	fooAddr := f.KeyAddress(helpers.KeyFoo)

	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, f.QueryBalances(fooAddr).AmountOf(sdk.DefaultBondDenom))

	proposalsQuery := f.QueryGovProposals()
	require.Empty(t, proposalsQuery)

	// Test submit generate only for submit proposal
	proposalTokens := sdk.TokensFromConsensusPower(5)
	success, stdout, stderr := f.TxGovSubmitProposal(
		fooAddr.String(), "Text", "Test", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "--generate-only", "-y")
	require.True(t, success)
	require.Empty(t, stderr)
	msg := helpers.UnmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = f.TxGovSubmitProposal(helpers.KeyFoo, "Text", "Test", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "--dry-run")
	require.True(t, success)

	// Create the proposal
	f.TxGovSubmitProposal(helpers.KeyFoo, "Text", "Test", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure transaction events can be queried
	searchResult := f.QueryTxs(1, 50, "message.action=submit_proposal", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure deposit was deducted
	require.Equal(t, startTokens.Sub(proposalTokens), f.QueryBalances(fooAddr).AmountOf(helpers.Denom))

	// Ensure propsal is directly queryable
	proposal1 := f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusDepositPeriod, proposal1.Status)

	// Ensure query proposals returns properly
	proposalsQuery = f.QueryGovProposals()
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// Query the deposits on the proposal
	deposit := f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(helpers.Denom))

	// Test deposit generate only
	depositTokens := sdk.TokensFromConsensusPower(10)
	success, stdout, stderr = f.TxGovDeposit(1, fooAddr.String(), sdk.NewCoin(helpers.Denom, depositTokens), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = helpers.UnmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Run the deposit transaction
	f.TxGovDeposit(1, helpers.KeyFoo, sdk.NewCoin(helpers.Denom, depositTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// test query deposit
	deposits := f.QueryGovDeposits(1)
	require.Len(t, deposits, 1)
	require.Equal(t, proposalTokens.Add(depositTokens), deposits[0].Amount.AmountOf(helpers.Denom))

	// Ensure querying the deposit returns the proper amount
	deposit = f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens.Add(depositTokens), deposit.Amount.AmountOf(helpers.Denom))

	// Ensure events are set on the transaction
	searchResult = f.QueryTxs(1, 50, "message.action=deposit", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure account has expected amount of funds
	require.Equal(t, startTokens.Sub(proposalTokens.Add(depositTokens)), f.QueryBalances(fooAddr).AmountOf(helpers.Denom))

	// Fetch the proposal and ensure it is now in the voting period
	proposal1 = f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusVotingPeriod, proposal1.Status)

	// Test vote generate only
	success, stdout, stderr = f.TxGovVote(1, gov.OptionYes, fooAddr.String(), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = helpers.UnmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Vote on the proposal
	f.TxGovVote(1, gov.OptionYes, helpers.KeyFoo, "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Query the vote
	vote := f.QueryGovVote(1, fooAddr)
	require.Equal(t, uint64(1), vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	// Query the votes
	votes := f.QueryGovVotes(1)
	require.Len(t, votes, 1)
	require.Equal(t, uint64(1), votes[0].ProposalID)
	require.Equal(t, gov.OptionYes, votes[0].Option)

	// Ensure events are applied to voting transaction properly
	searchResult = f.QueryTxs(1, 50, "message.action=vote", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure no proposals in deposit period
	proposalsQuery = f.QueryGovProposals("--status=DepositPeriod")
	require.Empty(t, proposalsQuery)

	// Ensure the proposal returns as in the voting period
	proposalsQuery = f.QueryGovProposals("--status=VotingPeriod")
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// submit a second test proposal
	f.TxGovSubmitProposal(helpers.KeyFoo, "Text", "Apples", "test", sdk.NewCoin(helpers.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Test limit on proposals query
	proposalsQuery = f.QueryGovProposals("--limit=2")
	require.Len(t, proposalsQuery, 2)
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	f.Cleanup()
}

func TestGaiaCLISubmitParamChangeProposal(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	proc := f.SDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(helpers.KeyFoo)
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, f.QueryBalances(fooAddr).AmountOf(sdk.DefaultBondDenom))

	// write proposal to file
	proposalTokens := sdk.TokensFromConsensusPower(5)
	proposal := fmt.Sprintf(`{
  "title": "Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 105
    }
  ],
  "deposit": "%sstake"
}
`, proposalTokens.String())

	proposalFile := helpers.WriteToNewTempFile(t, proposal)

	// create the param change proposal
	f.TxGovSubmitParamChangeProposal(helpers.KeyFoo, proposalFile.Name(), sdk.NewCoin(helpers.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure transaction events can be queried
	txsPage := f.QueryTxs(1, 50, "message.action=submit_proposal", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, txsPage.Txs, 1)

	// ensure deposit was deducted
	require.Equal(t, startTokens.Sub(proposalTokens).String(), f.QueryBalances(fooAddr).AmountOf(sdk.DefaultBondDenom).String())

	// ensure proposal is directly queryable
	proposal1 := f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusDepositPeriod, proposal1.Status)

	// ensure correct query proposals result
	proposalsQuery := f.QueryGovProposals()
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// ensure the correct deposit amount on the proposal
	deposit := f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(helpers.Denom))

	// Cleanup testing directories
	f.Cleanup()
}

func TestGaiaCLISubmitCommunityPoolSpendProposal(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// create some inflation
	genesisState := f.GenesisState()
	inflationMin := sdk.MustNewDecFromStr("1.0")
	var mintData mint.GenesisState
	cdc.UnmarshalJSON(genesisState[mint.ModuleName], &mintData)
	mintData.Minter.Inflation = inflationMin
	mintData.Params.InflationMin = inflationMin
	mintData.Params.InflationMax = sdk.MustNewDecFromStr("1.0")
	mintDataBz, err := cdc.MarshalJSON(mintData)
	require.NoError(t, err)
	genesisState[mint.ModuleName] = mintDataBz

	genFile := filepath.Join(f.SimdHome, "config", "genesis.json")
	genDoc, err := tmtypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	genDoc.AppState, err = cdc.MarshalJSON(genesisState)
	require.NoError(t, genDoc.SaveAs(genFile))

	proc := f.SDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(helpers.KeyFoo)
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, f.QueryBalances(fooAddr).AmountOf(sdk.DefaultBondDenom))

	tests.WaitForNextNBlocksTM(3, f.Port)

	// write proposal to file
	proposalTokens := sdk.TokensFromConsensusPower(5)
	proposal := fmt.Sprintf(`{
  "title": "Community Pool Spend",
  "description": "Spend from community pool",
  "recipient": "%s",
  "amount": "1%s",
  "deposit": "%s%s"
}
`, fooAddr, sdk.DefaultBondDenom, proposalTokens.String(), sdk.DefaultBondDenom)
	proposalFile := helpers.WriteToNewTempFile(t, proposal)

	// create the param change proposal
	f.TxGovSubmitCommunityPoolSpendProposal(helpers.KeyFoo, proposalFile.Name(), sdk.NewCoin(helpers.Denom, proposalTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure transaction events can be queried
	txsPage := f.QueryTxs(1, 50, "message.action=submit_proposal", fmt.Sprintf("message.sender=%s", fooAddr))
	require.Len(t, txsPage.Txs, 1)

	// ensure deposit was deducted
	require.Equal(t, startTokens.Sub(proposalTokens).String(), f.QueryBalances(fooAddr).AmountOf(sdk.DefaultBondDenom).String())

	// ensure proposal is directly queryable
	proposal1 := f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusDepositPeriod, proposal1.Status)

	// ensure correct query proposals result
	proposalsQuery := f.QueryGovProposals()
	require.Equal(t, uint64(1), proposalsQuery[0].ProposalID)

	// ensure the correct deposit amount on the proposal
	deposit := f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(helpers.Denom))

	// Cleanup testing directories
	f.Cleanup()
}
