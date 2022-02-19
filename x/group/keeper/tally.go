package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

func (k Keeper) Tally(ctx sdk.Context, proposalId uint64) (tallyResult group.TallyResult, err error) {
	proposal, err := k.getProposal(ctx, proposalId)
	if err != nil {
		return group.TallyResult{}, err
	}

	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED {
		return proposal.FinalTallyResult, nil
	}

	var policyInfo group.GroupPolicyInfo
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.Address); err != nil {
		return group.TallyResult{}, sdkerrors.Wrap(err, "load group policy")
	}

	// Ensure that group hasn't been modified since the proposal submission.
	electorate, err := k.getGroupInfo(ctx, policyInfo.GroupId)
	if err != nil {
		return group.TallyResult{}, err
	}
	if electorate.Version != proposal.GroupVersion {
		return group.TallyResult{}, sdkerrors.Wrap(errors.ErrModified, "group was modified")
	}

	votesIt, err := k.voteByVoterIndex.Get(ctx.KVStore(k.key), proposalId)
	if err != nil {
		return group.TallyResult{}, err
	}
	defer votesIt.Close()

	for {
		var vote group.Vote
		_, err := votesIt.LoadNext(&vote)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}

		voterAddr := vote.Voter
		voter := group.GroupMember{GroupId: policyInfo.GroupId, Member: &group.Member{Address: voterAddr}}
		if err := k.groupMemberTable.GetOne(ctx.KVStore(k.key), orm.PrimaryKey(&voter), &voter); err != nil {
			return group.TallyResult{}, sdkerrors.Wrapf(err, "address: %s", voterAddr)
		}

		err = tallyResult.Add(vote, voter.Member.Weight)
		if err != nil {
			return group.TallyResult{}, err
		}
	}

	return tallyResult, nil
}
