package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

func (q Keeper) Tally(ctx sdk.Context, p group.Proposal, groupId uint64) (group.TallyResult, error) {
	// If proposal has already been tallied and updated, then its status is
	// closed, in which case we just return the previously stored result.
	if p.Status == group.PROPOSAL_STATUS_CLOSED {
		return p.FinalTallyResult, nil
	}

	it, err := q.voteByProposalIndex.Get(ctx.KVStore(q.key), p.Id)
	if err != nil {
		return group.TallyResult{}, err
	}
	defer it.Close()

	tallyResult := group.DefaultTallyResult()

	var vote group.Vote
	for {
		_, err = it.LoadNext(&vote)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		if err != nil {
			return group.TallyResult{}, err
		}

		var member group.GroupMember
		err := q.groupMemberTable.GetOne(ctx.KVStore(q.key), orm.PrimaryKey(&group.GroupMember{GroupId: groupId, Member: &group.Member{Address: vote.Voter}}), &member)
		if err != nil {
			return group.TallyResult{}, err
		}

		if err := tallyResult.Add(vote, member.Member.Weight); err != nil {
			return group.TallyResult{}, sdkerrors.Wrap(err, "add new vote")
		}
	}

	return tallyResult, nil
}
