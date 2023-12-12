package v6

import (
	"cosmossdk.io/collections"
	v1 "cosmossdk.io/x/gov/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateStore performs in-place store migrations from v5 (v0.50) to v6 (v0.51). The
// migration includes:
//
// Addition of new field in params to store types of proposals that can be submitted.
func MigrateStore(ctx sdk.Context, proposalCollection collections.Map[uint64, v1.Proposal]) error {
	// Migrate proposals
	return proposalCollection.Walk(ctx, nil, func(key uint64, proposal v1.Proposal) (bool, error) {
		if proposal.Expedited {
			proposal.ProposalType = v1.ProposalType_PROPOSAL_TYPE_EXPEDITED
		} else {
			proposal.ProposalType = v1.ProposalType_PROPOSAL_TYPE_STANDARD
		}

		if err := proposalCollection.Set(ctx, key, proposal); err != nil {
			return false, err
		}

		return false, nil
	})
}
