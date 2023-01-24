package cli

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/require"
)

func TestParseArgsToContent(t *testing.T) {
	fs := NewCmdSubmitLegacyUpgradeProposal().Flags()

	proposal := types.SoftwareUpgradeProposal{ //nolint:staticcheck // SA1019: types.SoftwareUpgradeProposal is deprecated: use types.Content instead
		Title:       "proposal title",
		Description: "proposal description",
		Plan: types.Plan{
			Name:   "plan name",
			Height: 123456,
			Info:   "plan info",
		},
	}

	fs.Set(cli.FlagTitle, proposal.Title)
	fs.Set(cli.FlagDescription, proposal.Description) //nolint:staticcheck // SA1019: cli.FlagDescription is deprecated: use cli.FlagProposalDescription instead
	fs.Set(FlagUpgradeHeight, strconv.FormatInt(proposal.Plan.Height, 10))
	fs.Set(FlagUpgradeInfo, proposal.Plan.Info)

	content, err := parseArgsToContent(fs, proposal.Plan.Name)
	require.NoError(t, err)

	p, ok := content.(*types.SoftwareUpgradeProposal) //nolint:staticcheck // SA1019: types.SoftwareUpgradeProposal is deprecated: use types.Content instead
	require.Equal(t, ok, true)
	require.Equal(t, p.Title, proposal.Title)
	require.Equal(t, p.Description, proposal.Description)
	require.Equal(t, p.Plan.Name, proposal.Plan.Name)
	require.Equal(t, p.Plan.Height, proposal.Plan.Height)
	require.Equal(t, p.Plan.Info, proposal.Plan.Info)
}
