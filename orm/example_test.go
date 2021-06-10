package orm_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GroupKeeper struct {
	key                      sdk.StoreKey
	groupTable               orm.AutoUInt64Table
	groupByAdminIndex        orm.Index
	groupMemberTable         orm.PrimaryKeyTable
	groupMemberByGroupIndex  orm.Index
	groupMemberByMemberIndex orm.Index
}

var (
	GroupTablePrefix               byte = 0x0
	GroupTableSeqPrefix            byte = 0x1
	GroupByAdminIndexPrefix        byte = 0x2
	GroupMemberTablePrefix         byte = 0x3
	GroupMemberTableSeqPrefix      byte = 0x4
	GroupMemberTableIndexPrefix    byte = 0x5
	GroupMemberByGroupIndexPrefix  byte = 0x6
	GroupMemberByMemberIndexPrefix byte = 0x7
)

func NewGroupKeeper(storeKey sdk.StoreKey, cdc codec.Codec) GroupKeeper {
	k := GroupKeeper{key: storeKey}

	groupTableBuilder := orm.NewAutoUInt64TableBuilder(GroupTablePrefix, GroupTableSeqPrefix, storeKey, &testdata.GroupInfo{}, cdc)
	// note: quite easy to mess with Index prefixes when managed outside. no fail fast on duplicates
	k.groupByAdminIndex = orm.NewIndex(groupTableBuilder, GroupByAdminIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		return []orm.RowID{[]byte(val.(*testdata.GroupInfo).Admin)}, nil
	})
	k.groupTable = groupTableBuilder.Build()

	groupMemberTableBuilder := orm.NewPrimaryKeyTableBuilder(GroupMemberTablePrefix, storeKey, &testdata.GroupMember{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)

	k.groupMemberByGroupIndex = orm.NewIndex(groupMemberTableBuilder, GroupMemberByGroupIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		group := val.(*testdata.GroupMember).Group
		return []orm.RowID{[]byte(group)}, nil
	})
	k.groupMemberByMemberIndex = orm.NewIndex(groupMemberTableBuilder, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		return []orm.RowID{[]byte(val.(*testdata.GroupMember).Member)}, nil
	})
	k.groupMemberTable = groupMemberTableBuilder.Build()

	return k
}
