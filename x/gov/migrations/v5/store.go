package v5

import (
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// AddProposerAddressToProposal will add proposer to proposal
// and set to the store. This function is optional, and only needed
// if you wish that migrated proposals be cancellable.
func AddProposerAddressToProposal(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, proposals map[uint64]string) error {
	proposalIDs := make([]uint64, 0, len(proposals))

	for proposalID := range proposals {
		proposalIDs = append(proposalIDs, proposalID)
	}

	// sort the proposalIDs
	sort.Slice(proposalIDs, func(i, j int) bool { return proposalIDs[i] < proposalIDs[j] })

	store := ctx.KVStore(storeKey)

	for _, proposalID := range proposalIDs {
		if len(proposals[proposalID]) == 0 {
			return fmt.Errorf("found missing proposer for proposal ID: %d", proposalID)
		}

		if _, err := sdk.AccAddressFromBech32(proposals[proposalID]); err != nil {
			return fmt.Errorf("invalid proposer address : %s", proposals[proposalID])
		}

		bz := store.Get(types.ProposalKey(proposalID))
		var proposal govv1.Proposal
		if err := cdc.Unmarshal(bz, &proposal); err != nil {
			panic(err)
		}

		// Check if proposal is active
		if proposal.Status != govv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD &&
			proposal.Status != govv1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD {
			return fmt.Errorf("invalid proposal : %s, proposal not active", proposals[proposalID])
		}

		proposal.Proposer = proposals[proposalID]

		// set the new proposal with proposer
		bz, err := cdc.Marshal(&proposal)
		if err != nil {
			panic(err)
		}
		store.Set(types.ProposalKey(proposal.Id), bz)
	}

	return nil
}
