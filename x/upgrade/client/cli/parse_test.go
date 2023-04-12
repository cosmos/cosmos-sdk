package cli

import (
	"strconv"
	"testing"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/upgrade/types"
)

func TestParsePlan(t *testing.T) {
	fs := NewCmdSubmitUpgradeProposal(addresscodec.NewBech32Codec("cosmos")).Flags()

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
