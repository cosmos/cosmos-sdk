package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

const (
	// Group Table
	GroupTablePrefix        byte = 0x0
	GroupTableSeqPrefix     byte = 0x1
	GroupByAdminIndexPrefix byte = 0x2

	// Group Member Table
	GroupMemberTablePrefix         byte = 0x10
	GroupMemberByGroupIndexPrefix  byte = 0x11
	GroupMemberByMemberIndexPrefix byte = 0x12

	// Group Policy Table
	GroupPolicyTablePrefix        byte = 0x20
	GroupPolicyTableSeqPrefix     byte = 0x21
	GroupPolicyByGroupIndexPrefix byte = 0x22
	GroupPolicyByAdminIndexPrefix byte = 0x23

	// Proposal Table
	ProposalTablePrefix              byte = 0x30
	ProposalTableSeqPrefix           byte = 0x31
	ProposalByGroupPolicyIndexPrefix byte = 0x32
	ProposalsByVotingPeriodEndPrefix byte = 0x33

	// Vote Table
	VoteTablePrefix           byte = 0x40
	VoteByProposalIndexPrefix byte = 0x41
	VoteByVoterIndexPrefix    byte = 0x42
)

type Keeper struct {
	key storetypes.StoreKey

	accKeeper group.AccountKeeper

	// Group Table
	groupTable        orm.AutoUInt64Table
	groupByAdminIndex orm.Index

	// Group Member Table
	groupMemberTable         orm.PrimaryKeyTable
	groupMemberByGroupIndex  orm.Index
	groupMemberByMemberIndex orm.Index

	// Group Policy Table
	groupPolicySeq          orm.Sequence
	groupPolicyTable        orm.PrimaryKeyTable
	groupPolicyByGroupIndex orm.Index
	groupPolicyByAdminIndex orm.Index

	// Proposal Table
	proposalTable              orm.AutoUInt64Table
	proposalByGroupPolicyIndex orm.Index
	proposalsByVotingPeriodEnd orm.Index

	// Vote Table
	voteTable           orm.PrimaryKeyTable
	voteByProposalIndex orm.Index
	voteByVoterIndex    orm.Index

	router *authmiddleware.MsgServiceRouter

	config group.Config
}

// NewKeeper creates a new group keeper.
func NewKeeper(storeKey storetypes.StoreKey, cdc codec.Codec, router *authmiddleware.MsgServiceRouter, accKeeper group.AccountKeeper, config group.Config) Keeper {
	k := Keeper{
		key:       storeKey,
		router:    router,
		accKeeper: accKeeper,
	}

	groupTable, err := orm.NewAutoUInt64Table([2]byte{GroupTablePrefix}, GroupTableSeqPrefix, &group.GroupInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.groupByAdminIndex, err = orm.NewIndex(groupTable, GroupByAdminIndexPrefix, func(val interface{}) ([]interface{}, error) {
		addr, err := sdk.AccAddressFromBech32(val.(*group.GroupInfo).Admin)
		if err != nil {
			return nil, err
		}
		return []interface{}{addr.Bytes()}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.groupTable = *groupTable

	// Group Member Table
	groupMemberTable, err := orm.NewPrimaryKeyTable([2]byte{GroupMemberTablePrefix}, &group.GroupMember{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.groupMemberByGroupIndex, err = orm.NewIndex(groupMemberTable, GroupMemberByGroupIndexPrefix, func(val interface{}) ([]interface{}, error) {
		group := val.(*group.GroupMember).GroupId
		return []interface{}{group}, nil
	}, group.GroupMember{}.GroupId)
	if err != nil {
		panic(err.Error())
	}
	k.groupMemberByMemberIndex, err = orm.NewIndex(groupMemberTable, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]interface{}, error) {
		memberAddr := val.(*group.GroupMember).Member.Address
		addr, err := sdk.AccAddressFromBech32(memberAddr)
		if err != nil {
			return nil, err
		}
		return []interface{}{addr.Bytes()}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.groupMemberTable = *groupMemberTable

	// Group Policy Table
	k.groupPolicySeq = orm.NewSequence(GroupPolicyTableSeqPrefix)
	groupPolicyTable, err := orm.NewPrimaryKeyTable([2]byte{GroupPolicyTablePrefix}, &group.GroupPolicyInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.groupPolicyByGroupIndex, err = orm.NewIndex(groupPolicyTable, GroupPolicyByGroupIndexPrefix, func(value interface{}) ([]interface{}, error) {
		return []interface{}{value.(*group.GroupPolicyInfo).GroupId}, nil
	}, group.GroupPolicyInfo{}.GroupId)
	if err != nil {
		panic(err.Error())
	}
	k.groupPolicyByAdminIndex, err = orm.NewIndex(groupPolicyTable, GroupPolicyByAdminIndexPrefix, func(value interface{}) ([]interface{}, error) {
		admin := value.(*group.GroupPolicyInfo).Admin
		addr, err := sdk.AccAddressFromBech32(admin)
		if err != nil {
			return nil, err
		}
		return []interface{}{addr.Bytes()}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.groupPolicyTable = *groupPolicyTable

	// Proposal Table
	proposalTable, err := orm.NewAutoUInt64Table([2]byte{ProposalTablePrefix}, ProposalTableSeqPrefix, &group.Proposal{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.proposalByGroupPolicyIndex, err = orm.NewIndex(proposalTable, ProposalByGroupPolicyIndexPrefix, func(value interface{}) ([]interface{}, error) {
		account := value.(*group.Proposal).GroupPolicyAddress
		addr, err := sdk.AccAddressFromBech32(account)
		if err != nil {
			return nil, err
		}
		return []interface{}{addr.Bytes()}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.proposalsByVotingPeriodEnd, err = orm.NewIndex(proposalTable, ProposalsByVotingPeriodEndPrefix, func(value interface{}) ([]interface{}, error) {
		votingPeriodEnd := value.(*group.Proposal).VotingPeriodEnd
		return []interface{}{sdk.FormatTimeBytes(votingPeriodEnd)}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.proposalTable = *proposalTable

	// Vote Table
	voteTable, err := orm.NewPrimaryKeyTable([2]byte{VoteTablePrefix}, &group.Vote{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.voteByProposalIndex, err = orm.NewIndex(voteTable, VoteByProposalIndexPrefix, func(value interface{}) ([]interface{}, error) {
		return []interface{}{value.(*group.Vote).ProposalId}, nil
	}, group.Vote{}.ProposalId)
	if err != nil {
		panic(err.Error())
	}
	k.voteByVoterIndex, err = orm.NewIndex(voteTable, VoteByVoterIndexPrefix, func(value interface{}) ([]interface{}, error) {
		addr, err := sdk.AccAddressFromBech32(value.(*group.Vote).Voter)
		if err != nil {
			return nil, err
		}
		return []interface{}{addr.Bytes()}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.voteTable = *voteTable

	if config.MaxMetadataLen == 0 {
		config.MaxMetadataLen = group.DefaultConfig().MaxMetadataLen
	}
	if config.MaxExecutionPeriod == 0 {
		config.MaxExecutionPeriod = group.DefaultConfig().MaxExecutionPeriod
	}
	k.config = config

	return k

}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", group.ModuleName))
}

// GetGroupSequence returns the current value of the group table sequence
func (k Keeper) GetGroupSequence(ctx sdk.Context) uint64 {
	return k.groupTable.Sequence().CurVal(ctx.KVStore(k.key))
}

// iterateProposalsByVPEnd iterates over all proposals whose voting_period_end is after the `endTime` time argument.
func (k Keeper) iterateProposalsByVPEnd(ctx sdk.Context, endTime time.Time, cb func(proposal group.Proposal) (bool, error)) error {
	timeBytes := sdk.FormatTimeBytes(endTime)
	it, err := k.proposalsByVotingPeriodEnd.PrefixScan(ctx.KVStore(k.key), nil, timeBytes)

	if err != nil {
		return err
	}
	defer it.Close()

	for {
		// Important: this following line cannot outside the for loop.
		// It seems that when one unmarshals into the same `group.Proposal`
		// reference, then gogoproto somehow "adds" the new bytes to the old
		// object for some fields. When running simulations, for proposals with
		// each 1-2 proposers, after a couple of loop iterations we got to a
		// proposal with 60k+ proposers.
		// So we're declaring a local variable that gets GCed.
		//
		// Also see `x/group/types/proposal_test.go`, TestGogoUnmarshalProposal().
		var proposal group.Proposal
		_, err := it.LoadNext(&proposal)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		if err != nil {
			return err
		}

		stop, err := cb(proposal)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	return nil
}

// pruneProposal deletes a proposal from state.
func (k Keeper) pruneProposal(ctx sdk.Context, proposalID uint64) error {
	store := ctx.KVStore(k.key)

	err := k.proposalTable.Delete(store, proposalID)
	if err != nil {
		return err
	}

	k.Logger(ctx).Debug(fmt.Sprintf("Pruned proposal %d", proposalID))
	return nil
}

// updateProposalStatus iterates through all proposals by group policy index and updates proposal status
func (k Keeper) updateProposalStatus(ctx sdk.Context, groupPolicyAddr sdk.AccAddress) error {
	proposalIt, err := k.proposalByGroupPolicyIndex.Get(ctx.KVStore(k.key), groupPolicyAddr.Bytes())
	if err != nil {
		return err
	}
	defer proposalIt.Close()

	for {
		var proposalInfo group.Proposal
		_, err = proposalIt.LoadNext(&proposalInfo)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		if err != nil {
			return err
		}
		proposalInfo.Status = group.PROPOSAL_STATUS_ABORTED

		if err := k.proposalTable.Update(ctx.KVStore(k.key), proposalInfo.Id, &proposalInfo); err != nil {
			return err
		}
	}
	return nil
}

// pruneVotes prunes all votes for a proposal from state.
func (k Keeper) pruneVotes(ctx sdk.Context, proposalID uint64) error {
	store := ctx.KVStore(k.key)
	it, err := k.voteByProposalIndex.Get(store, proposalID)
	if err != nil {
		return err
	}
	defer it.Close()

	for {
		var vote group.Vote
		_, err = it.LoadNext(&vote)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		if err != nil {
			return err
		}

		err = k.voteTable.Delete(store, &vote)
		if err != nil {
			return err
		}
	}

	return nil
}

// PruneProposals prunes all proposals that are expired, i.e. whose
// `voting_period + max_execution_period` is greater than the current block
// time.
func (k Keeper) PruneProposals(ctx sdk.Context) error {
	err := k.iterateProposalsByVPEnd(ctx, ctx.BlockTime().Add(-k.config.MaxExecutionPeriod), func(proposal group.Proposal) (bool, error) {
		err := k.pruneProposal(ctx, proposal.Id)
		if err != nil {
			return true, err
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// TallyProposalsAtVPEnd iterates over all proposals whose voting period
// has ended, tallies their votes, prunes them, and updates the proposal's
// `FinalTallyResult` field.
func (k Keeper) TallyProposalsAtVPEnd(ctx sdk.Context) error {
	return k.iterateProposalsByVPEnd(ctx, ctx.BlockTime(), func(proposal group.Proposal) (bool, error) {
		policyInfo, err := k.getGroupPolicyInfo(ctx, proposal.GroupPolicyAddress)
		if err != nil {
			return true, sdkerrors.Wrap(err, "group policy")
		}

		electorate, err := k.getGroupInfo(ctx, policyInfo.GroupId)
		if err != nil {
			return true, sdkerrors.Wrap(err, "group")
		}

		proposalId := proposal.Id
		if proposal.Status == group.PROPOSAL_STATUS_ABORTED || proposal.Status == group.PROPOSAL_STATUS_WITHDRAWN {
			if err := k.pruneProposal(ctx, proposalId); err != nil {
				return true, err
			}
			if err := k.pruneVotes(ctx, proposalId); err != nil {
				return true, err
			}

		} else {
			err = k.doTallyAndUpdate(ctx, &proposal, electorate, policyInfo)
			if err != nil {
				return true, sdkerrors.Wrap(err, "doTallyAndUpdate")
			}

			if err := k.proposalTable.Update(ctx.KVStore(k.key), proposal.Id, &proposal); err != nil {
				return true, sdkerrors.Wrap(err, "proposal update")
			}
		}

		return false, nil
	})
}
