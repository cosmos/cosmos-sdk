package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/group"
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

	// Group Account Table
	GroupAccountTablePrefix        byte = 0x20
	GroupAccountTableSeqPrefix     byte = 0x21
	GroupAccountByGroupIndexPrefix byte = 0x22
	GroupAccountByAdminIndexPrefix byte = 0x23

	// Proposal Table
	ProposalTablePrefix               byte = 0x30
	ProposalTableSeqPrefix            byte = 0x31
	ProposalByGroupAccountIndexPrefix byte = 0x32
	ProposalByProposerIndexPrefix     byte = 0x33

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

	// Group Account Table
	groupAccountSeq          orm.Sequence
	groupAccountTable        orm.PrimaryKeyTable
	groupAccountByGroupIndex orm.Index
	groupAccountByAdminIndex orm.Index

	// Proposal Table
	proposalTable               orm.AutoUInt64Table
	proposalByGroupAccountIndex orm.Index
	proposalByProposerIndex     orm.Index

	// Vote Table
	voteTable           orm.PrimaryKeyTable
	voteByProposalIndex orm.Index
	voteByVoterIndex    orm.Index

	router *authmiddleware.MsgServiceRouter
}

func NewKeeper(storeKey storetypes.StoreKey, cdc codec.Codec, router *authmiddleware.MsgServiceRouter, accKeeper group.AccountKeeper) Keeper {
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

	// Group Account Table
	k.groupAccountSeq = orm.NewSequence(GroupAccountTableSeqPrefix)
	groupAccountTable, err := orm.NewPrimaryKeyTable([2]byte{GroupAccountTablePrefix}, &group.GroupAccountInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.groupAccountByGroupIndex, err = orm.NewIndex(groupAccountTable, GroupAccountByGroupIndexPrefix, func(value interface{}) ([]interface{}, error) {
		return []interface{}{value.(*group.GroupAccountInfo).GroupId}, nil
	}, group.GroupAccountInfo{}.GroupId)
	if err != nil {
		panic(err.Error())
	}
	k.groupAccountByAdminIndex, err = orm.NewIndex(groupAccountTable, GroupAccountByAdminIndexPrefix, func(value interface{}) ([]interface{}, error) {
		admin := value.(*group.GroupAccountInfo).Admin
		addr, err := sdk.AccAddressFromBech32(admin)
		if err != nil {
			return nil, err
		}
		return []interface{}{addr.Bytes()}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.groupAccountTable = *groupAccountTable

	// Proposal Table
	proposalTable, err := orm.NewAutoUInt64Table([2]byte{ProposalTablePrefix}, ProposalTableSeqPrefix, &group.Proposal{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.proposalByGroupAccountIndex, err = orm.NewIndex(proposalTable, ProposalByGroupAccountIndexPrefix, func(value interface{}) ([]interface{}, error) {
		account := value.(*group.Proposal).Address
		addr, err := sdk.AccAddressFromBech32(account)
		if err != nil {
			return nil, err
		}
		return []interface{}{addr.Bytes()}, nil
	}, []byte{})
	if err != nil {
		panic(err.Error())
	}
	k.proposalByProposerIndex, err = orm.NewIndex(proposalTable, ProposalByProposerIndexPrefix, func(value interface{}) ([]interface{}, error) {
		proposers := value.(*group.Proposal).Proposers
		r := make([]interface{}, len(proposers))
		for i := range proposers {
			addr, err := sdk.AccAddressFromBech32(proposers[i])
			if err != nil {
				return nil, err
			}
			r[i] = addr.Bytes()
		}
		return r, nil
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

	return k

}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", group.ModuleName))
}
