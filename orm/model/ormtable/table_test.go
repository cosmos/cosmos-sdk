package ormtable_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
	"pgregory.net/rapid"

	queryv1beta1 "github.com/cosmos/cosmos-sdk/api/cosmos/base/query/v1beta1"
	sdkerrors "github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
	"github.com/cosmos/cosmos-sdk/orm/model/kv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

func TestScenario(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	// first run tests with a split index-commitment store
	runTestScenario(t, table, testkv.NewSplitMemBackend())

	// now run tests with a shared index-commitment store

	// we're going to wrap this test in a debug store and save the decoded debug
	// messages, these will be checked against a golden file at the end of the
	// test. the golden file can be used for fine-grained debugging of kv-store
	// layout
	debugBuf := &strings.Builder{}
	store := testkv.NewDebugBackend(
		testkv.NewSharedMemBackend(),
		&testkv.EntryCodecDebugger{
			EntryCodec: table,
			Print:      func(s string) { debugBuf.WriteString(s + "\n") },
		},
	)

	runTestScenario(t, table, store)

	// we're going to store debug data in a golden file to make sure that
	// logical decoding works successfully
	// run `go test pkgname -test.update-golden` to update the golden file
	// see https://pkg.go.dev/gotest.tools/v3/golden for docs
	golden.Assert(t, debugBuf.String(), "test_scenario.golden")

	checkEncodeDecodeEntries(t, table, store.IndexStoreReader())
}

// check that the ormkv.Entry's decode and encode to the same bytes
func checkEncodeDecodeEntries(t *testing.T, table ormtable.Table, store kv.ReadonlyStore) {
	it, err := store.Iterator(nil, nil)
	assert.NilError(t, err)
	for it.Valid() {
		key := it.Key()
		value := it.Value()
		entry, err := table.DecodeEntry(key, value)
		assert.NilError(t, err)
		k, v, err := table.EncodeEntry(entry)
		assert.Assert(t, bytes.Equal(key, k), "%x %x %s", key, k, entry)
		assert.Assert(t, bytes.Equal(value, v), "%x %x %s", value, v, entry)
		it.Next()
	}
}

func runTestScenario(t *testing.T, table ormtable.Table, backend ormtable.Backend) {
	ctx := ormtable.WrapContextDefault(backend)

	// let's create 10 data items we'll use later and give them indexes
	data := []*testpb.ExampleTable{
		{U32: 4, I64: -2, Str: "abc", U64: 7},  // 0
		{U32: 4, I64: -2, Str: "abd", U64: 7},  // 1
		{U32: 4, I64: -1, Str: "abc", U64: 8},  // 2
		{U32: 5, I64: -2, Str: "abd", U64: 8},  // 3
		{U32: 5, I64: -2, Str: "abe", U64: 9},  // 4
		{U32: 7, I64: -2, Str: "abe", U64: 10}, // 5
		{U32: 7, I64: -1, Str: "abe", U64: 11}, // 6
		{U32: 8, I64: -4, Str: "abc", U64: 11}, // 7
		{U32: 8, I64: 1, Str: "abc", U64: 12},  // 8
		{U32: 8, I64: 1, Str: "abd", U64: 10},  // 9
	}

	// let's make a function to match what's in our iterator with what we
	// expect using indexes in the data array above
	assertIteratorItems := func(it ormtable.Iterator, xs ...int) {
		for _, i := range xs {
			assert.Assert(t, it.Next())
			msg, err := it.GetMessage()
			assert.NilError(t, err)
			//t.Logf("data[%d] %v == %v", i, data[i], msg)
			assert.DeepEqual(t, data[i], msg, protocmp.Transform())
		}
		// make sure the iterator is done
		assert.Assert(t, !it.Next())
	}

	// insert one record
	err := table.Insert(ctx, data[0])
	// trivial prefix query has one record
	it, err := table.Iterator(ctx)
	assert.NilError(t, err)
	assertIteratorItems(it, 0)

	// insert one record
	err = table.Insert(ctx, data[1])
	// trivial prefix query has two records
	it, err = table.Iterator(ctx)
	assert.NilError(t, err)
	assertIteratorItems(it, 0, 1)

	// insert the other records
	assert.NilError(t, err)
	for i := 2; i < len(data); i++ {
		err = table.Insert(ctx, data[i])
		assert.NilError(t, err)
	}

	// let's do a prefix query on the primary key
	it, err = table.Iterator(ctx, ormlist.Prefix(uint32(8)))
	assert.NilError(t, err)
	assertIteratorItems(it, 7, 8, 9)

	// let's try a reverse prefix query
	it, err = table.Iterator(ctx, ormlist.Prefix(uint32(4)), ormlist.Reverse())
	assert.NilError(t, err)
	defer it.Close()
	assertIteratorItems(it, 2, 1, 0)

	// let's try a range query
	it, err = table.Iterator(ctx,
		ormlist.Start(uint32(4), int64(-1)),
		ormlist.End(uint32(7)),
	)
	assert.NilError(t, err)
	defer it.Close()
	assertIteratorItems(it, 2, 3, 4, 5, 6)

	// and another range query
	it, err = table.Iterator(ctx,
		ormlist.Start(uint32(5), int64(-3)),
		ormlist.End(uint32(8), int64(1), "abc"),
	)
	assert.NilError(t, err)
	defer it.Close()
	assertIteratorItems(it, 3, 4, 5, 6, 7, 8)

	// now a reverse range query on a different index
	strU32Index := table.GetIndex("str,u32")
	assert.Assert(t, strU32Index != nil)
	it, err = strU32Index.Iterator(ctx,
		ormlist.Start("abc"),
		ormlist.End("abd"),
		ormlist.Reverse(),
	)
	assertIteratorItems(it, 9, 3, 1, 8, 7, 2, 0)

	// another prefix query forwards
	it, err = strU32Index.Iterator(ctx, ormlist.Prefix("abe", uint32(7)))
	assertIteratorItems(it, 5, 6)
	// and backwards
	it, err = strU32Index.Iterator(ctx, ormlist.Prefix("abc", uint32(4)), ormlist.Reverse())
	assertIteratorItems(it, 2, 0)

	// try an unique index
	u64StrIndex := table.GetUniqueIndex("u64,str")
	assert.Assert(t, u64StrIndex != nil)
	found, err := u64StrIndex.Has(ctx, uint64(12), "abc")
	assert.NilError(t, err)
	assert.Assert(t, found)
	var a testpb.ExampleTable
	found, err = u64StrIndex.Get(ctx, &a, uint64(12), "abc")
	assert.NilError(t, err)
	assert.Assert(t, found)
	assert.DeepEqual(t, data[8], &a, protocmp.Transform())

	// let's try paginating some stuff

	// first create a function to test what we got from pagination
	assertGotItems := func(items []proto.Message, xs ...int) {
		n := len(xs)
		assert.Equal(t, n, len(items))
		for i := 0; i < n; i++ {
			j := xs[i]
			//t.Logf("data[%d] %v == %v", j, data[j], items[i])
			assert.DeepEqual(t, data[j], items[i], protocmp.Transform())
		}
	}

	// now do some pagination
	var items []proto.Message
	onItem := func(item proto.Message) {
		items = append(items, item)
	}
	res, err := ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Limit:      4,
			CountTotal: true,
		}}, onItem)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 4, len(res.Cursors))
	assertGotItems(items, 0, 1, 2, 3)

	// read another page
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Key:   res.NextKey,
			Limit: 4,
		}}, onItem)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 4, len(res.Cursors))
	assertGotItems(items, 4, 5, 6, 7)

	// and the last page
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Key:   res.NextKey,
			Limit: 4,
		}}, onItem)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, !res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(items, 8, 9)

	// let's go backwards
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Limit:      2,
			CountTotal: true,
			Reverse:    true,
		}}, onItem)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(items, 9, 8)

	// a bit more
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Key:     res.NextKey,
			Limit:   2,
			Reverse: true,
		}}, onItem)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(items, 7, 6)

	// range query
	items = nil
	res, err = ormtable.Paginate(table, ctx,
		&ormtable.PaginationRequest{
			PageRequest: &queryv1beta1.PageRequest{
				Limit: 10,
			},
		},
		onItem,
		ormlist.Start(uint32(4), int64(-1), "abc"),
		ormlist.End(uint32(7), int64(-2), "abe"),
	)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, !res.HaveMore)
	assert.Equal(t, 4, len(res.Cursors))
	assertGotItems(items, 2, 3, 4, 5)

	// let's try an offset
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Limit:      2,
			CountTotal: true,
			Offset:     3,
		}}, onItem)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(items, 3, 4)

	// and reverse
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Limit:      3,
			CountTotal: true,
			Offset:     5,
			Reverse:    true,
		}}, onItem)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 3, len(res.Cursors))
	assertGotItems(items, 4, 3, 2)

	// now an offset that's slightly too big
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Limit:      1,
			CountTotal: true,
			Offset:     10,
		}}, onItem)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(items))
	assert.Assert(t, !res.HaveMore)
	assert.Equal(t, uint64(10), res.Total)

	// another offset that's too big
	items = nil
	res, err = ormtable.Paginate(table, ctx, &ormtable.PaginationRequest{
		PageRequest: &queryv1beta1.PageRequest{
			Limit:      1,
			CountTotal: true,
			Offset:     14,
		}}, onItem)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(items))
	assert.Assert(t, !res.HaveMore)
	assert.Equal(t, uint64(10), res.Total)

	// now let's update some things
	for i := 0; i < 5; i++ {
		data[i].U64 = data[i].U64 * 2
		data[i].Bz = []byte(data[i].Str)
		err = table.Update(ctx, data[i])
		assert.NilError(t, err)
	}
	it, err = table.Iterator(ctx)
	assert.NilError(t, err)
	// we should still get everything in the same order
	assertIteratorItems(it, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

	// let's use SAVE_MODE_DEFAULT and add something
	data = append(data, &testpb.ExampleTable{U32: 9})
	err = table.Save(ctx, data[10])
	assert.NilError(t, err)
	pkIndex := table.GetUniqueIndex("u32,i64,str")
	found, err = pkIndex.Get(ctx, &a, uint32(9), int64(0), "")
	assert.NilError(t, err)
	assert.Assert(t, found)
	assert.DeepEqual(t, data[10], &a, protocmp.Transform())
	// and update it
	data[10].B = true
	assert.NilError(t, table.Save(ctx, data[10]))
	found, err = pkIndex.Get(ctx, &a, uint32(9), int64(0), "")
	assert.NilError(t, err)
	assert.Assert(t, found)
	assert.DeepEqual(t, data[10], &a, protocmp.Transform())
	// and iterate
	it, err = table.Iterator(ctx)
	assert.NilError(t, err)
	assertIteratorItems(it, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// let's export and import JSON and use a read-only backend
	buf := &bytes.Buffer{}
	readBackend := ormtable.NewReadBackend(ormtable.ReadBackendOptions{
		CommitmentStoreReader: backend.CommitmentStoreReader(),
		IndexStoreReader:      backend.IndexStoreReader(),
	})
	assert.NilError(t, table.ExportJSON(ormtable.WrapContextDefault(readBackend), buf))
	assert.NilError(t, table.ValidateJSON(bytes.NewReader(buf.Bytes())))
	store2 := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	assert.NilError(t, table.ImportJSON(store2, bytes.NewReader(buf.Bytes())))
	assertTablesEqual(t, table, ctx, store2)

	// let's delete item 5
	key5 := []interface{}{uint32(7), int64(-2), "abe"}
	err = pkIndex.DeleteByKey(ctx, key5...)
	assert.NilError(t, err)
	// it should be gone
	found, err = pkIndex.Has(ctx, key5...)
	assert.NilError(t, err)
	assert.Assert(t, !found)
	// and missing from the iterator
	it, err = table.Iterator(ctx)
	assert.NilError(t, err)
	assertIteratorItems(it, 0, 1, 2, 3, 4, 6, 7, 8, 9, 10)
}

func TestRandomTableData(t *testing.T) {
	testTable(t, TableDataGen(testutil.GenA, 100).Example().(*TableData))
}

func testTable(t *testing.T, tableData *TableData) {
	for _, index := range tableData.table.Indexes() {
		indexModel := &IndexModel{
			TableData: tableData,
			index:     index.(TestIndex),
		}
		sort.Sort(indexModel)
		if _, ok := index.(ormtable.UniqueIndex); ok {
			testUniqueIndex(t, indexModel)
		}
		testIndex(t, indexModel)
	}
}

func testUniqueIndex(t *testing.T, model *IndexModel) {
	index := model.index.(ormtable.UniqueIndex)
	t.Logf("testing unique index %T %s", index, index.Fields())
	for i := 0; i < len(model.data); i++ {
		x := model.data[i]
		ks, _, err := index.(ormkv.IndexCodec).EncodeKeyFromMessage(x.ProtoReflect())
		assert.NilError(t, err)

		values := protoValuesToInterfaces(ks)

		found, err := index.Has(model.context, values...)
		assert.NilError(t, err)
		assert.Assert(t, found)

		msg := model.table.MessageType().New().Interface()
		found, err = index.Get(model.context, msg, values...)
		assert.NilError(t, err)
		assert.Assert(t, found)
		assert.DeepEqual(t, x, msg, protocmp.Transform())
	}
}

func testIndex(t *testing.T, model *IndexModel) {
	index := model.index
	if index.IsFullyOrdered() {
		t.Logf("testing index %T %s", index, index.Fields())

		it, err := model.index.Iterator(model.context)
		assert.NilError(t, err)
		checkIteratorAgainstSlice(t, it, model.data)

		it, err = model.index.Iterator(model.context, ormlist.Reverse())
		assert.NilError(t, err)
		checkIteratorAgainstSlice(t, it, reverseData(model.data))

		rapid.Check(t, func(t *rapid.T) {
			i := rapid.IntRange(0, len(model.data)-2).Draw(t, "i").(int)
			j := rapid.IntRange(i+1, len(model.data)-1).Draw(t, "j").(int)

			start, _, err := model.index.(ormkv.IndexCodec).EncodeKeyFromMessage(model.data[i].ProtoReflect())
			assert.NilError(t, err)
			end, _, err := model.index.(ormkv.IndexCodec).EncodeKeyFromMessage(model.data[j].ProtoReflect())
			assert.NilError(t, err)

			startVals := protoValuesToInterfaces(start)
			endVals := protoValuesToInterfaces(end)

			it, err = model.index.Iterator(model.context, ormlist.Start(startVals...), ormlist.End(endVals...))
			assert.NilError(t, err)
			checkIteratorAgainstSlice(t, it, model.data[i:j+1])

			it, err = model.index.Iterator(model.context, ormlist.Start(startVals...), ormlist.End(endVals...), ormlist.Reverse())
			assert.NilError(t, err)
			checkIteratorAgainstSlice(t, it, reverseData(model.data[i:j+1]))
		})
	} else {
		t.Logf("testing unordered index %T %s", index, index.Fields())

		// get all the data
		it, err := model.index.Iterator(model.context)
		assert.NilError(t, err)
		var data2 []proto.Message
		for it.Next() {
			msg, err := it.GetMessage()
			assert.NilError(t, err)
			data2 = append(data2, msg)
		}
		assert.Equal(t, len(model.data), len(data2))

		// sort it
		model2 := &IndexModel{
			TableData: &TableData{
				table:   model.table,
				data:    data2,
				context: model.context,
			},
			index: model.index,
		}
		sort.Sort(model2)

		// compare
		for i := 0; i < len(data2); i++ {
			assert.DeepEqual(t, model.data[i], data2[i], protocmp.Transform())
		}
	}

}

func reverseData(data []proto.Message) []proto.Message {
	n := len(data)
	reverse := make([]proto.Message, n)
	for i := 0; i < n; i++ {
		reverse[n-i-1] = data[i]
	}
	return reverse
}

func checkIteratorAgainstSlice(t assert.TestingT, iterator ormtable.Iterator, data []proto.Message) {
	i := 0
	for iterator.Next() {
		if i >= len(data) {
			for iterator.Next() {
				i++
			}
			t.Log(fmt.Sprintf("too many elements in iterator, len(data) = %d, i = %d", len(data), i))
			t.FailNow()
		}
		msg, err := iterator.GetMessage()
		assert.NilError(t, err)
		assert.DeepEqual(t, data[i], msg, protocmp.Transform())
		i++
	}
}

func TableDataGen(elemGen *rapid.Generator, n int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *TableData {
		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)
		message := elemGen.Draw(t, "message").(proto.Message)
		table, err := ormtable.Build(ormtable.Options{
			Prefix:      prefix,
			MessageType: message.ProtoReflect().Type(),
		})
		if err != nil {
			panic(err)
		}

		data := make([]proto.Message, n)
		store := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())

		for i := 0; i < n; {
			message = elemGen.Draw(t, fmt.Sprintf("message[%d]", i)).(proto.Message)
			err := table.Insert(store, message)
			if sdkerrors.IsOf(err, ormerrors.PrimaryKeyConstraintViolation, ormerrors.UniqueKeyViolation) {
				continue
			} else if err != nil {
				panic(err)
			}
			data[i] = message
			i++
		}

		return &TableData{
			data:    data,
			table:   table,
			context: store,
		}
	})
}

type TableData struct {
	table   ormtable.Table
	data    []proto.Message
	context context.Context
}

type IndexModel struct {
	*TableData
	index TestIndex
}

// TestIndex exposes methods that all index implementations expose publicly
// but on private structs because they are intended only to be used for testing.
type TestIndex interface {
	ormtable.Index

	// CompareKeys the two keys against the underlying IndexCodec, returning a
	// negative value if key1 is less than key2, 0 if they are equal, and a
	// positive value otherwise.
	CompareKeys(key1, key2 []protoreflect.Value) int

	// IsFullyOrdered returns true if all of the fields in the index are
	// considered "well-ordered" in terms of sorted iteration.
	IsFullyOrdered() bool
}

func (m *IndexModel) Len() int {
	return len(m.data)
}

func (m *IndexModel) Less(i, j int) bool {
	is, _, err := m.index.(ormkv.IndexCodec).EncodeKeyFromMessage(m.data[i].ProtoReflect())
	if err != nil {
		panic(err)
	}
	js, _, err := m.index.(ormkv.IndexCodec).EncodeKeyFromMessage(m.data[j].ProtoReflect())
	if err != nil {
		panic(err)
	}
	return m.index.CompareKeys(is, js) < 0
}

func (m *IndexModel) Swap(i, j int) {
	x := m.data[i]
	m.data[i] = m.data[j]
	m.data[j] = x
}

var _ sort.Interface = &IndexModel{}

func TestJSONExportImport(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)
	store := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())

	for i := 0; i < 100; {
		x := testutil.GenA.Example().(proto.Message)
		err = table.Insert(store, x)
		if sdkerrors.IsOf(err, ormerrors.PrimaryKeyConstraintViolation, ormerrors.UniqueKeyViolation) {
			continue
		} else {
			assert.NilError(t, err)
		}
		i++
	}

	buf := &bytes.Buffer{}
	assert.NilError(t, table.ExportJSON(store, buf))

	assert.NilError(t, table.ValidateJSON(bytes.NewReader(buf.Bytes())))

	store2 := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	assert.NilError(t, table.ImportJSON(store2, bytes.NewReader(buf.Bytes())))

	assertTablesEqual(t, table, store, store2)
}

func assertTablesEqual(t assert.TestingT, table ormtable.Table, ctx, ctx2 context.Context) {
	it, err := table.Iterator(ctx)
	assert.NilError(t, err)
	it2, err := table.Iterator(ctx2)
	assert.NilError(t, err)

	for {
		have := it.Next()
		have2 := it2.Next()
		assert.Equal(t, have, have2)
		if !have {
			break
		}

		msg1, err := it.GetMessage()
		assert.NilError(t, err)
		msg2, err := it.GetMessage()
		assert.NilError(t, err)

		assert.DeepEqual(t, msg1, msg2, protocmp.Transform())
	}
}

func protoValuesToInterfaces(ks []protoreflect.Value) []interface{} {
	values := make([]interface{}, len(ks))
	for i := 0; i < len(ks); i++ {
		values[i] = ks[i].Interface()
	}

	return values
}
