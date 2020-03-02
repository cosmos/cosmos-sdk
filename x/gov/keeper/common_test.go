package keeper_test

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	TestProposal = types.NewTextProposal("Test", "description")
)

// ProposalEqual checks if two proposals are equal (note: slow, for tests only)
func ProposalEqual(proposalA types.Proposal, proposalB types.Proposal) bool {
	return bytes.Equal(types.ModuleCdc.MustMarshalBinaryBare(proposalA),
		types.ModuleCdc.MustMarshalBinaryBare(proposalB))
}
