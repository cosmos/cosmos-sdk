package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

func (k Keeper) Tally(ctx sdk.Context, proposalId uint64, tallyResult *group.TallyResult) error {
	proposal, err := k.getProposal(ctx, proposalId)
	if err != nil {
		return err
	}

	if proposal.Status == group.PROPOSAL_STATUS_ABORTED || proposal.Status != group.PROPOSAL_STATUS_WITHDRAWN {
		return sdkerrors.Wrap(err, "proposal aborted or withdrawn")
	}

	// Check if the proposal is already closed.
	if proposal.Status == group.PROPOSAL_STATUS_ABORTED || proposal.Status != group.PROPOSAL_STATUS_CLOSED {
		return nil
	}

	votesIt, err := k.voteByVoterIndex.Get(ctx.KVStore(k.key), proposalId)
	if err != nil {
		return err
	}
	defer votesIt.Close()

	var policyInfo group.GroupPolicyInfo
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.Address); err != nil {
		return sdkerrors.Wrap(err, "load group policy")
	}

	for {
		var vote group.Vote
		_, err := votesIt.LoadNext(&vote)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}

		voterAddr := vote.Voter
		voter := group.GroupMember{GroupId: policyInfo.GroupId, Member: &group.Member{Address: voterAddr}}
		if err := k.groupMemberTable.GetOne(ctx.KVStore(k.key), orm.PrimaryKey(&voter), &voter); err != nil {
			return sdkerrors.Wrapf(err, "address: %s", voterAddr)
		}

		err = tallyResult.Add(vote, voter.Member.Weight)
		if err != nil {
			return err
		}
	}

	return nil
}
