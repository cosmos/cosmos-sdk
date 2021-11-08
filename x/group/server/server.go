package server

import (
	"github.com/cosmos/cosmos-sdk/codec"
	servermodule "github.com/cosmos/cosmos-sdk/types/module/server"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/exported"
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

type serverImpl struct {
	key servermodule.RootModuleKey

	accKeeper  exported.AccountKeeper
	bankKeeper exported.BankKeeper

	// Group Table
	groupSeq   orm.Sequence
	groupTable orm.AutoUInt64Table
	// groupByAdminIndex orm.Index

	// Group Member Table
	groupMemberTable orm.PrimaryKeyTable
	// groupMemberByGroupIndex  orm.UInt64Index
	// groupMemberByMemberIndex orm.Index

	// Group Account Table
	groupAccountSeq   orm.Sequence
	groupAccountTable orm.PrimaryKeyTable
	// groupAccountByGroupIndex orm.UInt64Index
	// groupAccountByAdminIndex orm.Index

	// Proposal Table
	proposalTable orm.AutoUInt64Table
	// proposalByGroupAccountIndex orm.Index
	// proposalByProposerIndex     orm.Index

	// Vote Table
	voteTable orm.PrimaryKeyTable
	// voteByProposalIndex orm.UInt64Index
	// voteByVoterIndex    orm.Index
}

func newServer(storeKey servermodule.RootModuleKey, accKeeper exported.AccountKeeper, bankKeeper exported.BankKeeper, cdc codec.Codec) serverImpl {
	s := serverImpl{key: storeKey, accKeeper: accKeeper, bankKeeper: bankKeeper}

	// Group Table
	groupTable, err := orm.NewAutoUInt64Table([2]byte{GroupTablePrefix}, GroupTableSeqPrefix, &group.GroupInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.groupSeq = orm.NewSequence(GroupTableSeqPrefix)
	s.groupTable = *groupTable

	// Group Member Table
	groupMemberTable, err := orm.NewPrimaryKeyTable([2]byte{GroupMemberTablePrefix}, &group.GroupMember{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.groupMemberTable = *groupMemberTable

	// Group Account Table
	s.groupAccountSeq = orm.NewSequence(GroupAccountTableSeqPrefix)
	groupAccountTable, err := orm.NewPrimaryKeyTable([2]byte{GroupAccountTablePrefix}, &group.GroupAccountInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.groupAccountTable = *groupAccountTable

	// Proposal Table
	proposalTable, err := orm.NewAutoUInt64Table([2]byte{ProposalTablePrefix}, ProposalTableSeqPrefix, &group.Proposal{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.proposalTable = *proposalTable

	// Vote Table
	voteTable, err := orm.NewPrimaryKeyTable([2]byte{VoteTablePrefix}, &group.Vote{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.voteTable = *voteTable

	return s
}

func RegisterServices(configurator servermodule.Configurator, accountKeeper exported.AccountKeeper, bankKeeper exported.BankKeeper) {
	impl := newServer(configurator.ModuleKey(), accountKeeper, bankKeeper, configurator.Marshaler())
	group.RegisterMsgServer(configurator.MsgServer(), impl)
	// group.RegisterQueryServer(configurator.QueryServer(), impl)
	// configurator.RegisterWeightedOperationsHandler(impl.WeightedOperations)

}
