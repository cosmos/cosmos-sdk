package v6

import (
	"fmt"

	"cosmossdk.io/collections"
	v1 "cosmossdk.io/x/gov/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateStore performs in-place store migrations from v5 (v0.50) to v6 (v0.51). The
// migration includes:
//
// Addition of new field in params to store types of proposals that can be submitted.
// Addition of gov params for optimistic proposals.
func MigrateStore(ctx sdk.Context, paramsCollection collections.Item[v1.Params], proposalCollection collections.Map[uint64, v1.Proposal]) error {
	// Migrate proposals
	err := proposalCollection.Walk(ctx, nil, func(key uint64, proposal v1.Proposal) (bool, error) {
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
	if err != nil {
		return err
	}

	// Migrate params
	govParams, err := paramsCollection.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gov params: %w", err)
	}

	defaultParams := v1.DefaultParams()
	govParams.OptimisticAuthorizedAddresses = defaultParams.OptimisticAuthorizedAddresses
	govParams.OptimisticRejectedThreshold = defaultParams.OptimisticRejectedThreshold
	govParams.ProposalCancelMaxPeriod = defaultParams.ProposalCancelMaxPeriod

	return paramsCollection.Set(ctx, govParams)
}
