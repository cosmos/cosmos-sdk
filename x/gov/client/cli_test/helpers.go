package test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/cli_test/helpers"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/stretchr/testify/require"
	"strings"
)

//___________________________________________________________________________________
// gaiacli query gov

// QueryGovParamDeposit is gaiacli query gov param deposit
func QueryGovParamDeposit(f *helpers.Fixtures) gov.DepositParams {
	cmd := fmt.Sprintf("%s query gov param deposit %s", f.SimcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var depositParam gov.DepositParams

	err := f.Cdc.UnmarshalJSON([]byte(out), &depositParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return depositParam
}

// QueryGovParamVoting is gaiacli query gov param voting
func QueryGovParamVoting(f *helpers.Fixtures) gov.VotingParams {
	cmd := fmt.Sprintf("%s query gov param voting %s", f.SimcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var votingParam gov.VotingParams

	err := f.Cdc.UnmarshalJSON([]byte(out), &votingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votingParam
}

// QueryGovParamTallying is gaiacli query gov param tallying
func QueryGovParamTallying(f *helpers.Fixtures) gov.TallyParams {
	cmd := fmt.Sprintf("%s query gov param tallying %s", f.SimcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var tallyingParam gov.TallyParams

	err := f.Cdc.UnmarshalJSON([]byte(out), &tallyingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return tallyingParam
}

// QueryGovProposal is gaiacli query gov proposal
func QueryGovProposal(f *helpers.Fixtures, proposalID int, flags ...string) gov.Proposal {
	cmd := fmt.Sprintf("%s query gov proposal %d %v", f.SimcliBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var proposal gov.Proposal

	err := f.Cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return proposal
}

// QueryGovProposals is gaiacli query gov proposals
func QueryGovProposals(f *helpers.Fixtures, flags ...string) gov.Proposals {
	cmd := fmt.Sprintf("%s query gov proposals %v", f.SimcliBinary, f.Flags())
	stdout, stderr := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	if strings.Contains(stderr, "no matching proposals found") {
		return gov.Proposals{}
	}
	require.Empty(f.T, stderr)
	var out gov.Proposals

	err := f.Cdc.UnmarshalJSON([]byte(stdout), &out)
	require.NoError(f.T, err)
	return out
}

// QueryGovVote is gaiacli query gov vote
func QueryGovVote(f *helpers.Fixtures, proposalID int, voter sdk.AccAddress, flags ...string) gov.Vote {
	cmd := fmt.Sprintf("%s query gov vote %d %s %v", f.SimcliBinary, proposalID, voter, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var vote gov.Vote

	err := f.Cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return vote
}

// QueryGovVotes is gaiacli query gov votes
func QueryGovVotes(f *helpers.Fixtures, proposalID int, flags ...string) []gov.Vote {
	cmd := fmt.Sprintf("%s query gov votes %d %v", f.SimcliBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var votes []gov.Vote

	err := f.Cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votes
}

// QueryGovDeposit is gaiacli query gov deposit
func QueryGovDeposit(f *helpers.Fixtures, proposalID int, depositor sdk.AccAddress, flags ...string) gov.Deposit {
	cmd := fmt.Sprintf("%s query gov deposit %d %s %v", f.SimcliBinary, proposalID, depositor, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var deposit gov.Deposit

	err := f.Cdc.UnmarshalJSON([]byte(out), &deposit)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposit
}

// QueryGovDeposits is gaiacli query gov deposits
func QueryGovDeposits(f *helpers.Fixtures, propsalID int, flags ...string) []gov.Deposit {
	cmd := fmt.Sprintf("%s query gov deposits %d %v", f.SimcliBinary, propsalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var deposits []gov.Deposit

	err := f.Cdc.UnmarshalJSON([]byte(out), &deposits)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposits
}

//___________________________________________________________________________________
// gaiacli tx gov

// TxGovSubmitProposal is gaiacli tx gov submit-proposal
func TxGovSubmitProposal(f *helpers.Fixtures, from, typ, title, description string, deposit sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov submit-proposal %v --keyring-backend=test --from=%s --type=%s",
		f.SimcliBinary, f.Flags(), from, typ)
	cmd += fmt.Sprintf(" --title=%s --description=%s --deposit=%s", title, description, deposit)
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovDeposit is gaiacli tx gov deposit
func TxGovDeposit(f *helpers.Fixtures, proposalID int, from string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov deposit %d %s --keyring-backend=test --from=%s %v",
		f.SimcliBinary, proposalID, amount, from, f.Flags())
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovVote is gaiacli tx gov vote
func TxGovVote(f *helpers.Fixtures, proposalID int, option gov.VoteOption, from string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov vote %d %s --keyring-backend=test --from=%s %v",
		f.SimcliBinary, proposalID, option, from, f.Flags())
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovSubmitParamChangeProposal executes a CLI parameter change proposal
// submission.
func TxGovSubmitParamChangeProposal(f *helpers.Fixtures,
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {

	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal param-change %s --keyring-backend=test --from=%s %v",
		f.SimcliBinary, proposalPath, from, f.Flags(),
	)

	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovSubmitCommunityPoolSpendProposal executes a CLI community pool spend proposal
// submission.
func TxGovSubmitCommunityPoolSpendProposal(f *helpers.Fixtures,
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {

	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal community-pool-spend %s --keyring-backend=test --from=%s %v",
		f.SimcliBinary, proposalPath, from, f.Flags(),
	)

	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}
