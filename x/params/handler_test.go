package params_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
	"github.com/cosmos/cosmos-sdk/x/params"
)

func testProposal(changes ...params.Change) params.ProposalChange {
	return params.NewProposalChange(
		"Test",
		"description",
		"myspace",
		changes,
	)
}
func TestProposalPassedEndblocker(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := gov.GetMockApp(t, 1, gov.GenesisState{}, nil)

	tp := testProposal(params.NewChange([]byte{0x00}, nil, "mychange"))
	resTags := gov.TestProposal(t, mapp, addrs[0], keeper, sk, tp)

	require.Equal(t, sdk.MakeTag(tags.ProposalResult, tags.ActionProposalPassed), resTags[1])
}
