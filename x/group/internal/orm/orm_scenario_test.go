package orm

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestKeeperEndToEndWithAutoUInt64Table(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := NewMockContext()
	store := ctx.KVStore(sdk.NewKVStoreKey("test"))

	k := NewTestKeeper(cdc)

	tm := testdata.TableModel{
		Id:       1,
		Name:     "name",
		Number:   123,
		Metadata: []byte("metadata"),
	}
	// when stored
	rowID, err := k.autoUInt64Table.Create(store, &tm)
	require.NoError(t, err)
	// then we should find it
	exists := k.autoUInt64Table.Has(store, rowID)
	require.True(t, exists)

	// and load it
	var loaded testdata.TableModel

	binKey, err := k.autoUInt64Table.GetOne(store, rowID, &loaded)
	require.NoError(t, err)

	require.Equal(t, rowID, binary.BigEndian.Uint64(binKey))
	require.Equal(t, tm, loaded)

	// and exists in MultiKeyIndex
	exists, err = k.autoUInt64TableModelByMetadataIndex.Has(store, []byte("metadata"))
	require.NoError(t, err)
	require.True(t, exists)

	// and when loaded
	it, err := k.autoUInt64TableModelByMetadataIndex.Get(store, []byte("metadata"))
	require.NoError(t, err)

	// then
	binKey, loaded = first(t, it)
	assert.Equal(t, rowID, binary.BigEndian.Uint64(binKey))
	assert.Equal(t, tm, loaded)

	// when updated
	tm.Metadata = []byte("new-metadata")
	err = k.autoUInt64Table.Update(store, rowID, &tm)
	require.NoError(t, err)

	binKey, err = k.autoUInt64Table.GetOne(store, rowID, &loaded)
	require.NoError(t, err)

	require.Equal(t, rowID, binary.BigEndian.Uint64(binKey))
	require.Equal(t, tm, loaded)

	// then indexes are updated, too
	exists, err = k.autoUInt64TableModelByMetadataIndex.Has(store, []byte("new-metadata"))
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = k.autoUInt64TableModelByMetadataIndex.Has(store, []byte("metadata"))
	require.NoError(t, err)
	require.False(t, exists)

	// when deleted
	err = k.autoUInt64Table.Delete(store, rowID)
	require.NoError(t, err)

	exists = k.autoUInt64Table.Has(store, rowID)
	require.False(t, exists)

	// and also removed from secondary MultiKeyIndex
	exists, err = k.autoUInt64TableModelByMetadataIndex.Has(store, []byte("new-metadata"))
	require.NoError(t, err)
	require.False(t, exists)
}

func TestKeeperEndToEndWithPrimaryKeyTable(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := NewMockContext()
	store := ctx.KVStore(sdk.NewKVStoreKey("test"))

	k := NewTestKeeper(cdc)

	tm := testdata.TableModel{
		Id:       1,
		Name:     "name",
		Number:   123,
		Metadata: []byte("metadata"),
	}
	// when stored
	err := k.primaryKeyTable.Create(store, &tm)
	require.NoError(t, err)
	// then we should find it by primary key
	primaryKey := PrimaryKey(&tm)
	exists := k.primaryKeyTable.Has(store, primaryKey)
	require.True(t, exists)

	// and load it by primary key
	var loaded testdata.TableModel
	err = k.primaryKeyTable.GetOne(store, primaryKey, &loaded)
	require.NoError(t, err)

	// then values should match expectations
	require.Equal(t, tm, loaded)

	// and then the data should exists in MultiKeyIndex
	exists, err = k.primaryKeyTableModelByNumberIndex.Has(store, tm.Number)
	require.NoError(t, err)
	require.True(t, exists)

	// and when loaded from MultiKeyIndex
	it, err := k.primaryKeyTableModelByNumberIndex.Get(store, tm.Number)
	require.NoError(t, err)

	// then values should match as before
	_, err = First(it, &loaded)
	require.NoError(t, err)
	assert.Equal(t, tm, loaded)

	// and when we create another entry with the same primary key
	err = k.primaryKeyTable.Create(store, &tm)
	// then it should fail as the primary key must be unique
	require.True(t, errors.ErrORMUniqueConstraint.Is(err), err)

	// and when entity updated with new primary key
	updatedMember := &testdata.TableModel{
		Id:       2,
		Name:     tm.Name,
		Number:   tm.Number,
		Metadata: tm.Metadata,
	}
	// then it should fail as the primary key is immutable
	err = k.primaryKeyTable.Update(store, updatedMember)
	require.Error(t, err)

	// and when entity updated with non primary key attribute modified
	updatedMember = &testdata.TableModel{
		Id:       1,
		Name:     "new name",
		Number:   tm.Number,
		Metadata: tm.Metadata,
	}
	// then it should not fail
	err = k.primaryKeyTable.Update(store, updatedMember)
	require.NoError(t, err)

	// and when entity deleted
	err = k.primaryKeyTable.Delete(store, &tm)
	require.NoError(t, err)

	// it is removed from primaryKeyTable
	exists = k.primaryKeyTable.Has(store, primaryKey)
	require.False(t, exists)

	// and removed from secondary MultiKeyIndex
	exists, err = k.primaryKeyTableModelByNumberIndex.Has(store, tm.Number)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestGasCostsPrimaryKeyTable(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := NewMockContext()
	store := ctx.KVStore(sdk.NewKVStoreKey("test"))

	k := NewTestKeeper(cdc)

	tm := testdata.TableModel{
		Id:       1,
		Name:     "name",
		Number:   123,
		Metadata: []byte("metadata"),
	}
	rowID, err := k.autoUInt64Table.Create(store, &tm)
	require.NoError(t, err)
	require.Equal(t, uint64(1), rowID)

	gCtx := NewGasCountingMockContext()
	err = k.primaryKeyTable.Create(gCtx.KVStore(store), &tm)
	require.NoError(t, err)
	t.Logf("gas consumed on create: %d", gCtx.GasConsumed())

	// get by primary key
	gCtx.ResetGasMeter()
	var loaded testdata.TableModel
	err = k.primaryKeyTable.GetOne(gCtx.KVStore(store), PrimaryKey(&tm), &loaded)
	require.NoError(t, err)
	t.Logf("gas consumed on get by primary key: %d", gCtx.GasConsumed())

	// get by secondary index
	gCtx.ResetGasMeter()
	// and when loaded from MultiKeyIndex
	it, err := k.primaryKeyTableModelByNumberIndex.Get(gCtx.KVStore(store), tm.Number)
	require.NoError(t, err)
	var loadedSlice []testdata.TableModel
	_, err = ReadAll(it, &loadedSlice)
	require.NoError(t, err)
	t.Logf("gas consumed on get by multi index key: %d", gCtx.GasConsumed())

	// delete
	gCtx.ResetGasMeter()
	err = k.primaryKeyTable.Delete(gCtx.KVStore(store), &tm)
	require.NoError(t, err)
	t.Logf("gas consumed on delete by primary key: %d", gCtx.GasConsumed())

	// with 3 elements
	var tms []testdata.TableModel
	for i := 1; i < 4; i++ {
		gCtx.ResetGasMeter()
		tm := testdata.TableModel{
			Id:       uint64(i),
			Name:     fmt.Sprintf("name%d", i),
			Number:   123,
			Metadata: []byte("metadata"),
		}
		err = k.primaryKeyTable.Create(gCtx.KVStore(store), &tm)
		require.NoError(t, err)
		t.Logf("%d: gas consumed on create: %d", i, gCtx.GasConsumed())
		tms = append(tms, tm)
	}

	for i := 1; i < 4; i++ {
		gCtx.ResetGasMeter()
		tm := testdata.TableModel{
			Id:       uint64(i),
			Name:     fmt.Sprintf("name%d", i),
			Number:   123,
			Metadata: []byte("metadata"),
		}
		err = k.primaryKeyTable.GetOne(gCtx.KVStore(store), PrimaryKey(&tm), &loaded)
		require.NoError(t, err)
		t.Logf("%d: gas consumed on get by primary key: %d", i, gCtx.GasConsumed())
	}

	// get by secondary index
	gCtx.ResetGasMeter()
	// and when loaded from MultiKeyIndex
	it, err = k.primaryKeyTableModelByNumberIndex.Get(gCtx.KVStore(store), tm.Number)
	require.NoError(t, err)
	_, err = ReadAll(it, &loadedSlice)
	require.NoError(t, err)
	require.Len(t, loadedSlice, 3)
	t.Logf("gas consumed on get by multi index key: %d", gCtx.GasConsumed())

	// delete
	for i, m := range tms {
		gCtx.ResetGasMeter()

		err = k.primaryKeyTable.Delete(gCtx.KVStore(store), &m)
		require.NoError(t, err)
		t.Logf("%d: gas consumed on delete: %d", i, gCtx.GasConsumed())
	}
}

func first(t *testing.T, it Iterator) ([]byte, testdata.TableModel) {
	var loaded testdata.TableModel
	key, err := First(it, &loaded)
	require.NoError(t, err)
	return key, loaded
}
