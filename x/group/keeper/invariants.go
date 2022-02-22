package keeper

import (
	"fmt"
	"math"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	groupmath "github.com/cosmos/cosmos-sdk/x/group/internal/math"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

const (
	votesInvariant    = "Tally-Votes"
	weightInvariant   = "Group-TotalWeight"
	votesSumInvariant = "Tally-Votes-Sum"
)

// RegisterInvariants registers all group invariants
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(group.ModuleName, votesInvariant, TallyVotesInvariant(keeper))
	ir.RegisterRoute(group.ModuleName, weightInvariant, GroupTotalWeightInvariant(keeper))
	ir.RegisterRoute(group.ModuleName, votesSumInvariant, TallyVotesSumInvariant(keeper))
}

// TallyVotesInvariant checks that vote tally sums must never have less than the block before.
func TallyVotesInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if ctx.BlockHeight()-1 < 0 {
			return sdk.FormatInvariant(group.ModuleName, votesInvariant, "Not enough blocks to perform TallyVotesInvariant"), false
		}
		prevCtx, _ := ctx.CacheContext()
		prevCtx = prevCtx.WithBlockHeight(ctx.BlockHeight() - 1)
		msg, broken := TallyVotesInvariantHelper(ctx, prevCtx, keeper.key, keeper.proposalTable)
		return sdk.FormatInvariant(group.ModuleName, votesInvariant, msg), broken
	}
}

// GroupTotalWeightInvariant checks that group's TotalWeight must be equal to the sum of its members.
func GroupTotalWeightInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broken := GroupTotalWeightInvariantHelper(ctx, keeper.key, keeper.groupTable, keeper.groupMemberByGroupIndex)
		return sdk.FormatInvariant(group.ModuleName, weightInvariant, msg), broken
	}
}

// TallyVotesSumInvariant checks that proposal FinalTallyResult must correspond to the vote option.
func TallyVotesSumInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broken := TallyVotesSumInvariantHelper(ctx, keeper.key, keeper.groupTable, keeper.proposalTable, keeper.groupMemberTable, keeper.voteByProposalIndex, keeper.groupPolicyTable)
		return sdk.FormatInvariant(group.ModuleName, votesSumInvariant, msg), broken
	}
}

func TallyVotesInvariantHelper(ctx sdk.Context, prevCtx sdk.Context, key storetypes.StoreKey, proposalTable orm.AutoUInt64Table) (string, bool) {

	var msg string
	var broken bool

	prevIt, err := proposalTable.PrefixScan(prevCtx.KVStore(key), 1, math.MaxUint64)
	if err != nil {
		msg += fmt.Sprintf("PrefixScan failure on proposal table at block height %d\n%v\n", prevCtx.BlockHeight(), err)
		return msg, broken
	}

	curIt, err := proposalTable.PrefixScan(ctx.KVStore(key), 1, math.MaxUint64)
	if err != nil {
		msg += fmt.Sprintf("PrefixScan failure on proposal table at block height %d\n%v\n", ctx.BlockHeight(), err)
		return msg, broken
	}

	var curProposals []*group.Proposal
	_, err = orm.ReadAll(curIt, &curProposals)
	if err != nil {
		msg += fmt.Sprintf("error while getting all the proposals at block height %d\n%v\n", ctx.BlockHeight(), err)
		return msg, broken
	}

	var prevProposals []*group.Proposal
	_, err = orm.ReadAll(prevIt, &prevProposals)
	if err != nil {
		msg += fmt.Sprintf("error while getting all the proposals at block height %d\n%v\n", prevCtx.BlockHeight(), err)
		return msg, broken
	}

	for i := 0; i < len(prevProposals); i++ {
		if prevProposals[i].Id == curProposals[i].Id {
			prevYesCount, err := prevProposals[i].FinalTallyResult.GetYesCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting yes votes weight of proposal at block height %d\n%v\n", prevCtx.BlockHeight(), err)
				return msg, broken
			}
			curYesCount, err := curProposals[i].FinalTallyResult.GetYesCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting yes votes weight of proposal at block height %d\n%v\n", ctx.BlockHeight(), err)
				return msg, broken
			}
			prevNoCount, err := prevProposals[i].FinalTallyResult.GetNoCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting no votes weight of proposal at block height %d\n%v\n", prevCtx.BlockHeight(), err)
				return msg, broken
			}
			curNoCount, err := curProposals[i].FinalTallyResult.GetNoCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting no votes weight of proposal at block height %d\n%v\n", ctx.BlockHeight(), err)
				return msg, broken
			}
			prevAbstainCount, err := prevProposals[i].FinalTallyResult.GetAbstainCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting abstain votes weight of proposal at block height %d\n%v\n", prevCtx.BlockHeight(), err)
				return msg, broken
			}
			curAbstainCount, err := curProposals[i].FinalTallyResult.GetAbstainCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting abstain votes weight of proposal at block height %d\n%v\n", ctx.BlockHeight(), err)
				return msg, broken
			}
			prevVetoCount, err := prevProposals[i].FinalTallyResult.GetNoWithVetoCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting veto votes weight of proposal at block height %d\n%v\n", prevCtx.BlockHeight(), err)
				return msg, broken
			}
			curVetoCount, err := curProposals[i].FinalTallyResult.GetNoWithVetoCount()
			if err != nil {
				msg += fmt.Sprintf("error while getting veto votes weight of proposal at block height %d\n%v\n", ctx.BlockHeight(), err)
				return msg, broken
			}
			if (curYesCount.Cmp(prevYesCount) == -1) || (curNoCount.Cmp(prevNoCount) == -1) || (curAbstainCount.Cmp(prevAbstainCount) == -1) || (curVetoCount.Cmp(prevVetoCount) == -1) {
				broken = true
				msg += "vote tally sums must never have less than the block before\n"
				return msg, broken
			}
		}
	}
	return msg, broken
}

func GroupTotalWeightInvariantHelper(ctx sdk.Context, key storetypes.StoreKey, groupTable orm.AutoUInt64Table, groupMemberByGroupIndex orm.Index) (string, bool) {

	var msg string
	var broken bool

	var groupInfo group.GroupInfo
	var groupMember group.GroupMember

	groupIt, err := groupTable.PrefixScan(ctx.KVStore(key), 1, math.MaxUint64)
	if err != nil {
		msg += fmt.Sprintf("PrefixScan failure on group table\n%v\n", err)
		return msg, broken
	}
	defer groupIt.Close()

	for {
		membersWeight, err := groupmath.NewNonNegativeDecFromString("0")
		if err != nil {
			msg += fmt.Sprintf("error while parsing positive dec zero for group member\n%v\n", err)
			return msg, broken
		}
		_, err = groupIt.LoadNext(&groupInfo)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		memIt, err := groupMemberByGroupIndex.Get(ctx.KVStore(key), groupInfo.Id)
		if err != nil {
			msg += fmt.Sprintf("error while returning group member iterator for group with ID %d\n%v\n", groupInfo.Id, err)
			return msg, broken
		}

		for {
			_, err = memIt.LoadNext(&groupMember)
			if errors.ErrORMIteratorDone.Is(err) {
				break
			}
			curMemWeight, err := groupmath.NewNonNegativeDecFromString(groupMember.GetMember().GetWeight())
			if err != nil {
				msg += fmt.Sprintf("error while parsing non-nengative decimal for group member %s\n%v\n", groupMember.Member.Address, err)
				return msg, broken
			}
			membersWeight, err = groupmath.Add(membersWeight, curMemWeight)
			if err != nil {
				msg += fmt.Sprintf("decimal addition error while adding group member voting weight to total voting weight\n%v\n", err)
				return msg, broken
			}
		}
		memIt.Close()

		groupWeight, err := groupmath.NewNonNegativeDecFromString(groupInfo.GetTotalWeight())
		if err != nil {
			msg += fmt.Sprintf("error while parsing non-nengative decimal for group with ID %d\n%v\n", groupInfo.Id, err)
			return msg, broken
		}

		if groupWeight.Cmp(membersWeight) != 0 {
			broken = true
			msg += fmt.Sprintf("group's TotalWeight must be equal to the sum of its members' weights\ngroup weight: %s\nSum of group members weights: %s\n", groupWeight.String(), membersWeight.String())
			break
		}
	}
	return msg, broken
}

func TallyVotesSumInvariantHelper(ctx sdk.Context, key storetypes.StoreKey, groupTable orm.AutoUInt64Table, proposalTable orm.AutoUInt64Table, groupMemberTable orm.PrimaryKeyTable, voteByProposalIndex orm.Index, groupPolicyTable orm.PrimaryKeyTable) (string, bool) {
	var msg string
	var broken bool

	var groupInfo group.GroupInfo
	var proposal group.Proposal
	var groupPolicy group.GroupPolicyInfo
	var groupMem group.GroupMember
	var vote group.Vote

	proposalIt, err := proposalTable.PrefixScan(ctx.KVStore(key), 1, math.MaxUint64)
	if err != nil {
		fmt.Println(err)
		msg += fmt.Sprintf("PrefixScan failure on proposal table\n%v\n", err)
		return msg, broken
	}
	defer proposalIt.Close()

	for {

		totalVotingWeight, err := groupmath.NewNonNegativeDecFromString("0")
		if err != nil {
			msg += fmt.Sprintf("error while parsing positive dec zero for total voting weight\n%v\n", err)
			return msg, broken
		}
		yesVoteWeight, err := groupmath.NewNonNegativeDecFromString("0")
		if err != nil {
			msg += fmt.Sprintf("error while parsing positive dec zero for yes voting weight\n%v\n", err)
			return msg, broken
		}
		noVoteWeight, err := groupmath.NewNonNegativeDecFromString("0")
		if err != nil {
			msg += fmt.Sprintf("error while parsing positive dec zero for no voting weight\n%v\n", err)
			return msg, broken
		}
		abstainVoteWeight, err := groupmath.NewNonNegativeDecFromString("0")
		if err != nil {
			msg += fmt.Sprintf("error while parsing positive dec zero for abstain voting weight\n%v\n", err)
			return msg, broken
		}
		vetoVoteWeight, err := groupmath.NewNonNegativeDecFromString("0")
		if err != nil {
			msg += fmt.Sprintf("error while parsing positive dec zero for veto voting weight\n%v\n", err)
			return msg, broken
		}

		_, err = proposalIt.LoadNext(&proposal)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}

		err = groupPolicyTable.GetOne(ctx.KVStore(key), orm.PrimaryKey(&group.GroupPolicyInfo{Address: proposal.Address}), &groupPolicy)
		if err != nil {
			msg += fmt.Sprintf("group policy not found for address: %s\n%v\n", proposal.Address, err)
			return msg, broken
		}

		if proposal.GroupPolicyVersion != groupPolicy.Version {
			msg += fmt.Sprintf("group policy with address %s was modified\n", groupPolicy.Address)
			return msg, broken
		}

		_, err = groupTable.GetOne(ctx.KVStore(key), groupPolicy.GroupId, &groupInfo)
		if err != nil {
			msg += fmt.Sprintf("group info not found for group id %d\n%v\n", groupPolicy.GroupId, err)
			return msg, broken
		}

		if groupInfo.Version != proposal.GroupVersion {
			msg += fmt.Sprintf("group with id %d was modified\n", groupInfo.Id)
			return msg, broken
		}

		voteIt, err := voteByProposalIndex.Get(ctx.KVStore(key), proposal.Id)
		if err != nil {
			msg += fmt.Sprintf("error while returning vote iterator for proposal with ID %d\n%v\n", proposal.Id, err)
			return msg, broken
		}

		for {
			_, err := voteIt.LoadNext(&vote)
			if errors.ErrORMIteratorDone.Is(err) {
				break
			}

			err = groupMemberTable.GetOne(ctx.KVStore(key), orm.PrimaryKey(&group.GroupMember{GroupId: groupPolicy.GroupId, Member: &group.Member{Address: vote.Voter}}), &groupMem)
			if err != nil {
				msg += fmt.Sprintf("group member not found with group ID %d and group member %s\n%v\n", groupPolicy.GroupId, vote.Voter, err)
				return msg, broken
			}

			curMemVotingWeight, err := groupmath.NewNonNegativeDecFromString(groupMem.Member.Weight)
			if err != nil {
				msg += fmt.Sprintf("error while parsing non-negative decimal for group member %s\n%v\n", groupMem.Member.Address, err)
				return msg, broken
			}
			totalVotingWeight, err = groupmath.Add(totalVotingWeight, curMemVotingWeight)
			if err != nil {
				msg += fmt.Sprintf("decimal addition error while adding current member voting weight to total voting weight\n%v\n", err)
				return msg, broken
			}

			switch vote.Option {
			case group.VOTE_OPTION_YES:
				yesVoteWeight, err = groupmath.Add(yesVoteWeight, curMemVotingWeight)
				if err != nil {
					msg += fmt.Sprintf("decimal addition error\n%v\n", err)
					return msg, broken
				}
			case group.VOTE_OPTION_NO:
				noVoteWeight, err = groupmath.Add(noVoteWeight, curMemVotingWeight)
				if err != nil {
					msg += fmt.Sprintf("decimal addition error\n%v\n", err)
					return msg, broken
				}
			case group.VOTE_OPTION_ABSTAIN:
				abstainVoteWeight, err = groupmath.Add(abstainVoteWeight, curMemVotingWeight)
				if err != nil {
					msg += fmt.Sprintf("decimal addition error\n%v\n", err)
					return msg, broken
				}
			case group.VOTE_OPTION_NO_WITH_VETO:
				vetoVoteWeight, err = groupmath.Add(vetoVoteWeight, curMemVotingWeight)
				if err != nil {
					msg += fmt.Sprintf("decimal addition error\n%v\n", err)
					return msg, broken
				}
			}
		}
		voteIt.Close()

		totalProposalVotes, err := proposal.FinalTallyResult.TotalCounts()
		if err != nil {
			msg += fmt.Sprintf("error while getting total weighted votes of proposal with ID %d\n%v\n", proposal.Id, err)
			return msg, broken
		}
		proposalYesCount, err := proposal.FinalTallyResult.GetYesCount()
		if err != nil {
			msg += fmt.Sprintf("error while getting the weighted sum of yes votes for proposal with ID %d\n%v\n", proposal.Id, err)
			return msg, broken
		}
		proposalNoCount, err := proposal.FinalTallyResult.GetNoCount()
		if err != nil {
			msg += fmt.Sprintf("error while getting the weighted sum of no votes for proposal with ID %d\n%v\n", proposal.Id, err)
			return msg, broken
		}
		proposalAbstainCount, err := proposal.FinalTallyResult.GetAbstainCount()
		if err != nil {
			msg += fmt.Sprintf("error while getting the weighted sum of abstain votes for proposal with ID %d\n%v\n", proposal.Id, err)
			return msg, broken
		}
		proposalVetoCount, err := proposal.FinalTallyResult.GetNoWithVetoCount()
		if err != nil {
			msg += fmt.Sprintf("error while getting the weighted sum of veto votes for proposal with ID %d\n%v\n", proposal.Id, err)
			return msg, broken
		}

		if totalProposalVotes.Cmp(totalVotingWeight) != 0 {
			broken = true
			msg += fmt.Sprintf("proposal FinalTallyResult must correspond to the sum of votes weights\nProposal with ID %d has total proposal votes %s, but got sum of votes weights %s\n", proposal.Id, totalProposalVotes.String(), totalVotingWeight.String())
			break
		}

		if (yesVoteWeight.Cmp(proposalYesCount) != 0) || (noVoteWeight.Cmp(proposalNoCount) != 0) || (abstainVoteWeight.Cmp(proposalAbstainCount) != 0) || (vetoVoteWeight.Cmp(proposalVetoCount) != 0) {
			broken = true
			msg += fmt.Sprintf("proposal FinalTallyResult must correspond to the vote option\nProposal with ID %d and voter address %s must correspond to the vote option\n", proposal.Id, vote.Voter)
			break
		}
	}
	return msg, broken
}
