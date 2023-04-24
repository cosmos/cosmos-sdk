package keeper

import (
	groupv1 "cosmossdk.io/api/cosmos/group/v1"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/orm/types/ormerrors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

// Tally is a function that tallies a proposal by iterating through its votes,
// and returns the tally result without modifying the proposal or any state.
func (k Keeper) Tally(ctx sdk.Context, p group.Proposal, groupID uint64) (group.TallyResult, error) {
	// If proposal has already been tallied and updated, then its status is
	// accepted/rejected, in which case we just return the previously stored result.
	//
	// In all other cases (including withdrawn, aborted...) we do the tally
	// again.
	if p.Status == group.PROPOSAL_STATUS_ACCEPTED || p.Status == group.PROPOSAL_STATUS_REJECTED {
		return p.FinalTallyResult, nil
	}

	it, err := k.state.VoteTable().List(ctx, groupv1.VoteProposalIdVoterIndexKey{}.WithProposalId(p.Id))
	if err != nil {
		return group.TallyResult{}, err
	}
	defer it.Close()

	tallyResult := group.DefaultTallyResult()

	for it.Next() {
		vote, err := it.Value()
		if err != nil {
			return group.TallyResult{}, err
		}

		member, err := k.state.GroupMemberTable().Get(ctx, groupID, vote.Voter)
		switch {
		case ormerrors.IsNotFound(err):
			// If the member left the group aformerrors.IsNotFound(err)ter voting, then we simply skip the
			// vote.
			continue
		case err != nil:
			// For any other errors, we stop and return the error.
			return group.TallyResult{}, err
		}

		if err := tallyResult.Add(group.VoteFromPulsar(vote), member.Member.Weight); err != nil {
			return group.TallyResult{}, errorsmod.Wrap(err, "add new vote")
		}
	}

	return tallyResult, nil
}
