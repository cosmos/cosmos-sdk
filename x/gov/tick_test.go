package gov

import (
	"testing"
)

func TestTallyPass(t *testing.T) {

	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewProposal(ctx, "Test", "description", "Text")
	proposalID := proposal.ProposalID
	keeper.SetProposal(ctx, proposal)

	gotProposal := keeper.GetProposal(ctx, proposalID)

}
