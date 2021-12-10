package ormtable

import (
	"bytes"
	"os"
	"testing"

	"google.golang.org/protobuf/testing/protocmp"

	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/memkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
)

func TestScenario(t *testing.T) {
	table, err := Build(TableOptions{
		MessageType: (&testpb.A{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)
	store := memkv.NewMemIndexCommitmentStore()

	// let's create 10 data items we'll use later and give them indexes
	data := []*testpb.A{
		{U32: 4, I64: -2, Str: "abc", U64: 7},  // 0
		{U32: 4, I64: -2, Str: "abd", U64: 7},  // 1
		{U32: 4, I64: -1, Str: "abc", U64: 8},  // 2
		{U32: 5, I64: -2, Str: "abd", U64: 8},  // 3
		{U32: 5, I64: -2, Str: "abe", U64: 9},  // 4
		{U32: 7, I64: -2, Str: "abe", U64: 9},  // 5
		{U32: 7, I64: -1, Str: "abe", U64: 10}, // 6
		{U32: 8, I64: -4, Str: "abc", U64: 10}, // 7
		{U32: 8, I64: 1, Str: "abc", U64: 10},  // 8
		{U32: 8, I64: 1, Str: "abd", U64: 10},  // 9
	}

	// insert the data
	for _, datum := range data {
		err = table.Save(store, datum, SAVE_MODE_INSERT)
		assert.NilError(t, err)
	}

	// let's make a function to match what's in our iterator with what we
	// expect using indexes in the data array above
	assertGotItems := func(it Iterator, xs ...int) {
		for _, i := range xs {
			assert.Assert(t, it.Next())
			msg, err := it.GetMessage()
			assert.NilError(t, err)
			t.Logf("data[%d] %v == %v", i, data[i], msg)
			assert.DeepEqual(t, data[i], msg, protocmp.Transform())
		}
		// make sure the iterator is done
		assert.Assert(t, !it.Next())
	}

	// let's do a prefix query on the primary key
	it, err := table.PrefixIterator(store, testutil.ValuesOf(uint32(8)), IteratorOptions{})
	assert.NilError(t, err)
	assertGotItems(it, 7, 8, 9)

	// let's try a reverse prefix query
	it, err = table.PrefixIterator(store, testutil.ValuesOf(uint32(4)), IteratorOptions{Reverse: true})
	assert.NilError(t, err)
	assertGotItems(it, 2, 1, 0)

	// let's try a range query
	it, err = table.RangeIterator(store, testutil.ValuesOf(uint32(4), int64(-1)), testutil.ValuesOf(uint32(7)), IteratorOptions{})
	assert.NilError(t, err)
	assertGotItems(it, 2, 3, 4, 5, 6)

	// and another range query
	it, err = table.RangeIterator(store, testutil.ValuesOf(uint32(5), int64(-3)), testutil.ValuesOf(uint32(8), int64(1), "abc"), IteratorOptions{})
	assert.NilError(t, err)
	assertGotItems(it, 3, 4, 5, 6, 7, 8)

	// now a range query on a different index
}

func TestExportImport(t *testing.T) {
	table, err := Build(TableOptions{
		MessageType: (&testpb.A{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)
	store := memkv.NewMemIndexCommitmentStore()

	for i := 0; i < 100; i++ {
		x := testutil.GenA.Example().(proto.Message)
		err = table.Save(store, x, SAVE_MODE_DEFAULT)
		assert.NilError(t, err)
	}

	buf := &bytes.Buffer{}
	assert.NilError(t, table.ExportJSON(store, buf))

	store2 := memkv.NewMemIndexCommitmentStore()
	assert.NilError(t, table.ImportJSON(store2, bytes.NewReader(buf.Bytes())))

	it, err := table.PrefixIterator(store, nil, IteratorOptions{})
	assert.NilError(t, err)
	it2, err := table.PrefixIterator(store2, nil, IteratorOptions{})
	assert.NilError(t, err)

	for {
		have := it.Next()
		have2 := it2.Next()
		assert.Equal(t, have, have2)
		if !have {
			break
		}

		panic("TODO: compare keys & values")
	}
}

func TestDumpJSON(t *testing.T) {
	table, err := Build(TableOptions{
		MessageType: (&testpb.A{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)
	store := memkv.NewMemIndexCommitmentStore()
	for i := 0; i < 100; i++ {
		x := testutil.GenA.Example().(proto.Message)
		err = table.Save(store, x, SAVE_MODE_DEFAULT)
		assert.NilError(t, err)
	}
	f, err := os.OpenFile("test_data.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	assert.NilError(t, err)
	err = table.ExportJSON(store, f)
	assert.NilError(t, err)
}
