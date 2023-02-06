package cli

import (
	"strconv"
	"testing"

	"cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"
)

func TestParsePlan(t *testing.T) {
	fs := NewCmdSubmitUpgradeProposal().Flags()

	proposal := types.MsgSoftwareUpgrade{
		Plan: types.Plan{
			Name:   "plan name",
			Height: 123456,
			Info:   "plan info",
		},
	}

	fs.Set(FlagUpgradeHeight, strconv.FormatInt(proposal.Plan.Height, 10))
	fs.Set(FlagUpgradeInfo, proposal.Plan.Info)

	p, err := parsePlan(fs, proposal.Plan.Name)
	require.NoError(t, err)
	require.Equal(t, p.Name, proposal.Plan.Name)
	require.Equal(t, p.Height, proposal.Plan.Height)
	require.Equal(t, p.Info, proposal.Plan.Info)
}
