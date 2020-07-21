package testutil

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/require"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

//___________________________________________________________________________________
// simcli query gov

// QueryGovParamDeposit is simcli query gov param deposit
func QueryGovParamDeposit(f *cli.Fixtures) types.DepositParams {
	cmd := fmt.Sprintf("%s query gov param deposit %s", f.SimdBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var depositParam types.DepositParams

	err := f.Cdc.UnmarshalJSON([]byte(out), &depositParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return depositParam
}

// QueryGovParamVoting is simcli query gov param voting
func QueryGovParamVoting(f *cli.Fixtures) types.VotingParams {
	cmd := fmt.Sprintf("%s query gov param voting %s", f.SimdBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var votingParam types.VotingParams

	err := f.Cdc.UnmarshalJSON([]byte(out), &votingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votingParam
}

// QueryGovParamTallying is simcli query gov param tallying
func QueryGovParamTallying(f *cli.Fixtures) types.TallyParams {
	cmd := fmt.Sprintf("%s query gov param tallying %s", f.SimdBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var tallyingParam types.TallyParams

	err := f.Cdc.UnmarshalJSON([]byte(out), &tallyingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return tallyingParam
}

// QueryGovProposal is simcli query gov proposal
func QueryGovProposal(f *cli.Fixtures, proposalID int, flags ...string) types.Proposal {
	cmd := fmt.Sprintf("%s query gov proposal %d %v", f.SimdBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	var proposal types.Proposal

	err := f.Cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return proposal
}

// QueryGovProposals is simcli query gov proposals
func QueryGovProposals(f *cli.Fixtures, flags ...string) types.Proposals {
	cmd := fmt.Sprintf("%s query gov proposals %v", f.SimdBinary, f.Flags())
	stdout, stderr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	if strings.Contains(stderr, "no matching proposals found") {
		return types.Proposals{}
	}
	require.Empty(f.T, stderr)
	var out types.Proposals

	err := f.Cdc.UnmarshalJSON([]byte(stdout), &out)
	require.NoError(f.T, err)
	return out
}

// QueryGovVote is simcli query gov vote
func QueryGovVote(f *cli.Fixtures, proposalID int, voter sdk.AccAddress, flags ...string) types.Vote {
	cmd := fmt.Sprintf("%s query gov vote %d %s %v", f.SimdBinary, proposalID, voter, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	var vote types.Vote

	err := f.Cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return vote
}

// QueryGovVotes is simcli query gov votes
func QueryGovVotes(f *cli.Fixtures, proposalID int, flags ...string) []types.Vote {
	cmd := fmt.Sprintf("%s query gov votes %d %v", f.SimdBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	var votes []types.Vote

	err := f.Cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votes
}

// QueryGovDeposit is simcli query gov deposit
func QueryGovDeposit(f *cli.Fixtures, proposalID int, depositor sdk.AccAddress, flags ...string) types.Deposit {
	cmd := fmt.Sprintf("%s query gov deposit %d %s %v", f.SimdBinary, proposalID, depositor, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	var deposit types.Deposit

	err := f.Cdc.UnmarshalJSON([]byte(out), &deposit)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposit
}

// QueryGovDeposits is simcli query gov deposits
func QueryGovDeposits(f *cli.Fixtures, propsalID int, flags ...string) []types.Deposit {
	cmd := fmt.Sprintf("%s query gov deposits %d %v", f.SimdBinary, propsalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	var deposits []types.Deposit

	err := f.Cdc.UnmarshalJSON([]byte(out), &deposits)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposits
}

//___________________________________________________________________________________
// simcli tx gov

// TxGovSubmitProposal is simcli tx gov submit-proposal
func TxGovSubmitProposal(f *cli.Fixtures, from, typ, title, description string, deposit sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov submit-proposal %v --keyring-backend=test --from=%s --type=%s",
		f.SimdBinary, f.Flags(), from, typ)
	cmd += fmt.Sprintf(" --title=%s --description=%s --deposit=%s", title, description, deposit)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovDeposit is simcli tx gov deposit
func TxGovDeposit(f *cli.Fixtures, proposalID int, from string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov deposit %d %s --keyring-backend=test --from=%s %v",
		f.SimdBinary, proposalID, amount, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovVote is simcli tx gov vote
func TxGovVote(f *cli.Fixtures, proposalID int, option types.VoteOption, from string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov vote %d %s --keyring-backend=test --from=%s %v",
		f.SimdBinary, proposalID, option, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovSubmitParamChangeProposal executes a CLI parameter change proposal
// submission.
func TxGovSubmitParamChangeProposal(f *cli.Fixtures,
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {

	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal param-change %s --keyring-backend=test --from=%s %v",
		f.SimdBinary, proposalPath, from, f.Flags(),
	)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovSubmitCommunityPoolSpendProposal executes a CLI community pool spend proposal
// submission.
func TxGovSubmitCommunityPoolSpendProposal(f *cli.Fixtures,
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {

	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal community-pool-spend %s --keyring-backend=test --from=%s %v",
		f.SimdBinary, proposalPath, from, f.Flags(),
	)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}
