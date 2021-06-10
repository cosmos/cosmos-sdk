package orm_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Testing ORM with arbitrary address length
const addrLen = 10

func TestKeeperEndToEndWithAutoUInt64Table(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	ctx := orm.NewMockContext()

	k := NewGroupKeeper(storeKey, cdc)

	g := testdata.GroupInfo{
		GroupId:     1,
		Description: "my test",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}
	// when stored
	rowID, err := k.groupTable.Create(ctx, &g)
	require.NoError(t, err)
	// then we should find it
	exists := k.groupTable.Has(ctx, rowID)
	require.True(t, exists)

	// and load it
	var loaded testdata.GroupInfo

	binKey, err := k.groupTable.GetOne(ctx, rowID, &loaded)
	require.NoError(t, err)

	assert.Equal(t, rowID, binary.BigEndian.Uint64(binKey))
	assert.Equal(t, "my test", loaded.Description)
	assert.Equal(t, sdk.AccAddress([]byte("admin-address")), loaded.Admin)

	// and exists in MultiKeyIndex
	exists = k.groupByAdminIndex.Has(ctx, []byte("admin-address"))
	require.True(t, exists)

	// and when loaded
	it, err := k.groupByAdminIndex.Get(ctx, []byte("admin-address"))
	require.NoError(t, err)

	// then
	binKey, loaded = first(t, it)
	assert.Equal(t, rowID, binary.BigEndian.Uint64(binKey))
	assert.Equal(t, g, loaded)

	// when updated
	g.Admin = []byte("new-admin-address")
	err = k.groupTable.Save(ctx, rowID, &g)
	require.NoError(t, err)

	// then indexes are updated, too
	exists = k.groupByAdminIndex.Has(ctx, []byte("new-admin-address"))
	require.True(t, exists)

	exists = k.groupByAdminIndex.Has(ctx, []byte("admin-address"))
	require.False(t, exists)

	// when deleted
	err = k.groupTable.Delete(ctx, rowID)
	require.NoError(t, err)

	// then removed from primary MultiKeyIndex
	exists = k.groupTable.Has(ctx, rowID)
	require.False(t, exists)

	// and also removed from secondary MultiKeyIndex
	exists = k.groupByAdminIndex.Has(ctx, []byte("new-admin-address"))
	require.False(t, exists)
}

func TestKeeperEndToEndWithPrimaryKeyTable(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	ctx := orm.NewMockContext()

	k := NewGroupKeeper(storeKey, cdc)

	g := testdata.GroupInfo{
		GroupId:     1,
		Description: "my test",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}

	m := testdata.GroupMember{
		Group:  sdk.AccAddress(orm.EncodeSequence(1)),
		Member: sdk.AccAddress([]byte("member-address")),
		Weight: 10,
	}
	groupRowID, err := k.groupTable.Create(ctx, &g)
	require.NoError(t, err)
	require.Equal(t, uint64(1), groupRowID)
	// when stored
	err = k.groupMemberTable.Create(ctx, &m)
	require.NoError(t, err)

	// then we should find it by primary key
	primaryKey := m.PrimaryKey()
	exists := k.groupMemberTable.Has(ctx, primaryKey)
	require.True(t, exists)
	// and load it by primary key
	var loaded testdata.GroupMember
	err = k.groupMemberTable.GetOne(ctx, primaryKey, &loaded)
	require.NoError(t, err)

	// then values should match expectations
	require.Equal(t, m, loaded)

	// and then the data should exists in MultiKeyIndex
	exists = k.groupMemberByGroupIndex.Has(ctx, orm.EncodeSequence(groupRowID))
	require.True(t, exists)

	// and when loaded from MultiKeyIndex
	it, err := k.groupMemberByGroupIndex.Get(ctx, orm.EncodeSequence(groupRowID))
	require.NoError(t, err)

	// then values should match as before
	_, err = orm.First(it, &loaded)
	require.NoError(t, err)

	assert.Equal(t, m, loaded)
	// and when we create another entry with the same primary key
	err = k.groupMemberTable.Create(ctx, &m)
	// then it should fail as the primary key must be unique
	require.True(t, orm.ErrUniqueConstraint.Is(err), err)

	// and when entity updated with new primary key
	updatedMember := &testdata.GroupMember{
		Group:  m.Group,
		Member: []byte("new-member-address"),
		Weight: m.Weight,
	}
	// then it should fail as the primary key is immutable
	err = k.groupMemberTable.Save(ctx, updatedMember)
	require.Error(t, err)

	// and when entity updated with non primary key attribute modified
	updatedMember = &testdata.GroupMember{
		Group:  m.Group,
		Member: m.Member,
		Weight: 99,
	}
	// then it should not fail
	err = k.groupMemberTable.Save(ctx, updatedMember)
	require.NoError(t, err)

	// and when entity deleted
	err = k.groupMemberTable.Delete(ctx, &m)
	require.NoError(t, err)

	// then it is removed from primary key MultiKeyIndex
	exists = k.groupMemberTable.Has(ctx, primaryKey)
	require.False(t, exists)

	// and removed from secondary MultiKeyIndex
	exists = k.groupMemberByGroupIndex.Has(ctx, orm.EncodeSequence(groupRowID))
	require.False(t, exists)
}

func TestGasCostsPrimaryKeyTable(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	ctx := orm.NewMockContext()

	k := NewGroupKeeper(storeKey, cdc)

	g := testdata.GroupInfo{
		GroupId:     1,
		Description: "my test",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}

	m := testdata.GroupMember{
		Group:  sdk.AccAddress(orm.EncodeSequence(1)),
		Member: sdk.AccAddress([]byte("member-address")),
		Weight: 10,
	}
	groupRowID, err := k.groupTable.Create(ctx, &g)
	require.NoError(t, err)
	require.Equal(t, uint64(1), groupRowID)
	gCtx := orm.NewGasCountingMockContext(ctx)
	err = k.groupMemberTable.Create(gCtx, &m)
	require.NoError(t, err)
	t.Logf("gas consumed on create: %d", gCtx.GasConsumed())

	// get by primary key
	gCtx.ResetGasMeter()
	var loaded testdata.GroupMember
	err = k.groupMemberTable.GetOne(gCtx, m.PrimaryKey(), &loaded)
	require.NoError(t, err)
	t.Logf("gas consumed on get by primary key: %d", gCtx.GasConsumed())

	// get by secondary index
	gCtx.ResetGasMeter()
	// and when loaded from MultiKeyIndex
	it, err := k.groupMemberByGroupIndex.Get(gCtx, orm.EncodeSequence(groupRowID))
	require.NoError(t, err)
	var loadedSlice []testdata.GroupMember
	_, err = orm.ReadAll(it, &loadedSlice)
	require.NoError(t, err)

	t.Logf("gas consumed on get by multi index key: %d", gCtx.GasConsumed())

	// delete
	gCtx.ResetGasMeter()
	err = k.groupMemberTable.Delete(gCtx, &m)
	require.NoError(t, err)
	t.Logf("gas consumed on delete by primary key: %d", gCtx.GasConsumed())

	// with 3 elements
	for i := 1; i < 4; i++ {
		gCtx.ResetGasMeter()
		m := testdata.GroupMember{
			Group:  sdk.AccAddress(orm.EncodeSequence(1)),
			Member: sdk.AccAddress([]byte(fmt.Sprintf("member-address%d", i))),
			Weight: 10,
		}
		err = k.groupMemberTable.Create(gCtx, &m)
		require.NoError(t, err)
		t.Logf("%d: gas consumed on create: %d", i, gCtx.GasConsumed())
	}

	for i := 1; i < 4; i++ {
		gCtx.ResetGasMeter()
		m := testdata.GroupMember{
			Group:  sdk.AccAddress(orm.EncodeSequence(1)),
			Member: sdk.AccAddress([]byte(fmt.Sprintf("member-address%d", i))),
			Weight: 10,
		}
		err = k.groupMemberTable.GetOne(gCtx, m.PrimaryKey(), &loaded)
		require.NoError(t, err)
		t.Logf("%d: gas consumed on get by primary key: %d", i, gCtx.GasConsumed())
	}

	// get by secondary index
	gCtx.ResetGasMeter()
	// and when loaded from MultiKeyIndex
	it, err = k.groupMemberByGroupIndex.Get(gCtx, orm.EncodeSequence(groupRowID))
	require.NoError(t, err)
	_, err = orm.ReadAll(it, &loadedSlice)
	require.NoError(t, err)
	require.Len(t, loadedSlice, 3)
	t.Logf("gas consumed on get by multi index key: %d", gCtx.GasConsumed())

	// delete
	for i, m := range loadedSlice {
		gCtx.ResetGasMeter()

		err = k.groupMemberTable.Delete(gCtx, &m)
		require.NoError(t, err)
		t.Logf("%d: gas consumed on delete: %d", i, gCtx.GasConsumed())
	}
}

func TestExportImportStateAutoUInt64Table(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	ctx := orm.NewMockContext()

	k := NewGroupKeeper(storeKey, cdc)

	testRecords := 10
	for i := 1; i <= testRecords; i++ {
		myAddr := sdk.AccAddress(bytes.Repeat([]byte{byte(i)}, addrLen))
		g := testdata.GroupInfo{
			GroupId:     uint64(i),
			Description: fmt.Sprintf("my test %d", i),
			Admin:       myAddr,
		}

		groupRowID, err := k.groupTable.Create(ctx, &g)
		require.NoError(t, err)
		require.Equal(t, uint64(i), groupRowID)
	}
	var groups []*testdata.GroupInfo
	seqVal, err := orm.ExportTableData(ctx, k.groupTable, &groups)
	require.NoError(t, err)

	// when a new db seeded
	ctx = orm.NewMockContext()

	err = orm.ImportTableData(ctx, k.groupTable, groups, seqVal)
	require.NoError(t, err)
	// then all data is set again

	for i := 1; i <= testRecords; i++ {
		require.True(t, k.groupTable.Has(ctx, uint64(i)))
		var loaded testdata.GroupInfo
		groupRowID, err := k.groupTable.GetOne(ctx, uint64(i), &loaded)
		require.NoError(t, err)

		require.Equal(t, orm.RowID(orm.EncodeSequence(uint64(i))), groupRowID)
		assert.Equal(t, fmt.Sprintf("my test %d", i), loaded.Description)
		exp := sdk.AccAddress(bytes.Repeat([]byte{byte(i)}, addrLen))
		assert.Equal(t, exp, loaded.Admin)

		// and also the indexes
		require.True(t, k.groupByAdminIndex.Has(ctx, exp))
		it, err := k.groupByAdminIndex.Get(ctx, exp)
		require.NoError(t, err)
		var all []testdata.GroupInfo
		orm.ReadAll(it, &all)
		require.Len(t, all, 1)
		assert.Equal(t, loaded, all[0])
	}
	require.Equal(t, uint64(testRecords), k.groupTable.Sequence().CurVal(ctx))
}

func TestExportImportStatePrimaryKeyTable(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	ctx := orm.NewMockContext()

	k := NewGroupKeeper(storeKey, cdc)
	myGroupAddr := sdk.AccAddress(bytes.Repeat([]byte{byte('a')}, addrLen))
	testRecordsNum := 10
	testRecords := make([]testdata.GroupMember, testRecordsNum)
	for i := 1; i <= testRecordsNum; i++ {
		myAddr := sdk.AccAddress(bytes.Repeat([]byte{byte(i)}, addrLen))
		g := testdata.GroupMember{
			Group:  myGroupAddr,
			Member: myAddr,
			Weight: uint64(i),
		}
		err := k.groupMemberTable.Create(ctx, &g)
		require.NoError(t, err)
		testRecords[i-1] = g
	}
	var groupMembers []*testdata.GroupMember
	_, err := orm.ExportTableData(ctx, k.groupMemberTable, &groupMembers)
	require.NoError(t, err)

	// when a new db seeded
	ctx = orm.NewMockContext()

	err = orm.ImportTableData(ctx, k.groupMemberTable, groupMembers, 0)
	require.NoError(t, err)

	// then all data is set again
	it, err := k.groupMemberTable.PrefixScan(ctx, nil, nil)
	require.NoError(t, err)
	var loaded []testdata.GroupMember
	keys, err := orm.ReadAll(it, &loaded)
	require.NoError(t, err)
	for i := range keys {
		assert.Equal(t, testRecords[i].PrimaryKey(), keys[i].Bytes())
	}
	assert.Equal(t, testRecords, loaded)

	// and first index setup
	it, err = k.groupMemberByGroupIndex.Get(ctx, myGroupAddr)
	require.NoError(t, err)
	loaded = nil
	keys, err = orm.ReadAll(it, &loaded)
	require.NoError(t, err)
	for i := range keys {
		assert.Equal(t, testRecords[i].PrimaryKey(), keys[i].Bytes())
	}
	assert.Equal(t, testRecords, loaded)

	// and second index setup
	for _, v := range testRecords {
		it, err = k.groupMemberByMemberIndex.Get(ctx, v.Member)
		require.NoError(t, err)
		loaded = nil
		keys, err = orm.ReadAll(it, &loaded)
		require.NoError(t, err)
		assert.Equal(t, []orm.RowID{v.PrimaryKey()}, keys)
		assert.Equal(t, []testdata.GroupMember{v}, loaded)
	}
}

func first(t *testing.T, it orm.Iterator) ([]byte, testdata.GroupInfo) {
	var loaded testdata.GroupInfo
	key, err := orm.First(it, &loaded)
	require.NoError(t, err)
	return key, loaded
}
