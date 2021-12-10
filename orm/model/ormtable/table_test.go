package ormtable

import (
	"bytes"
	"fmt"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/internal/memkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	assertIteratorItems := func(it Iterator, xs ...int) {
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

	// let's do a prefix query on the primary key
	it, err := table.PrefixIterator(store, testutil.ValuesOf(uint32(8)), IteratorOptions{})
	assert.NilError(t, err)
	assertIteratorItems(it, 7, 8, 9)

	// let's try a reverse prefix query
	it, err = table.PrefixIterator(store, testutil.ValuesOf(uint32(4)), IteratorOptions{Reverse: true})
	defer it.Close()
	assert.NilError(t, err)
	assertIteratorItems(it, 2, 1, 0)

	// let's try a range query
	it, err = table.RangeIterator(store, testutil.ValuesOf(uint32(4), int64(-1)), testutil.ValuesOf(uint32(7)), IteratorOptions{})
	defer it.Close()
	assert.NilError(t, err)
	assertIteratorItems(it, 2, 3, 4, 5, 6)

	// and another range query
	it, err = table.RangeIterator(store, testutil.ValuesOf(uint32(5), int64(-3)), testutil.ValuesOf(uint32(8), int64(1), "abc"), IteratorOptions{})
	defer it.Close()
	assert.NilError(t, err)
	assertIteratorItems(it, 3, 4, 5, 6, 7, 8)

	// now a reverse range query on a different index
	u64StrIndexFields, err := CommaSeparatedFieldNames("u64,str")
	assert.NilError(t, err)
	u64StrIndex := table.GetIndex(u64StrIndexFields)
	assert.Assert(t, u64StrIndex != nil)
	it, err = u64StrIndex.RangeIterator(store, testutil.ValuesOf(uint64(8)), testutil.ValuesOf(uint64(9)), IteratorOptions{Reverse: true})
	assertIteratorItems(it, 5, 4, 3, 2)

	// another prefix query forwards
	it, err = u64StrIndex.PrefixIterator(store, testutil.ValuesOf(uint64(10), "abc"), IteratorOptions{})
	assertIteratorItems(it, 7, 8)
	// and backwards
	it, err = u64StrIndex.PrefixIterator(store, testutil.ValuesOf(uint64(10), "abc"), IteratorOptions{Reverse: true})
	assertIteratorItems(it, 8, 7)

	// try an unique index
	strU32I64Fields, err := CommaSeparatedFieldNames("str,u32,i64")
	assert.NilError(t, err)
	strU32I64Index := table.GetUniqueIndex(strU32I64Fields)
	assert.Assert(t, strU32I64Index != nil)
	found, err := strU32I64Index.Has(store, testutil.ValuesOf("abc", uint32(8), int64(1)))
	assert.NilError(t, err)
	assert.Assert(t, found)
	var a testpb.A
	found, err = strU32I64Index.Get(store, testutil.ValuesOf("abc", uint32(8), int64(1)), &a)
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
	res, err := Paginate(table, store, &PaginationRequest{
		PageRequest: &query.PageRequest{
			Limit:      4,
			CountTotal: true,
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 4, len(res.Cursors))
	assertGotItems(res.Items, 0, 1, 2, 3)

	// read another page
	res, err = Paginate(table, store, &PaginationRequest{
		PageRequest: &query.PageRequest{
			Key:   res.NextKey,
			Limit: 4,
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 4, len(res.Cursors))
	assertGotItems(res.Items, 4, 5, 6, 7)

	// and the last page
	res, err = Paginate(table, store, &PaginationRequest{
		PageRequest: &query.PageRequest{
			Key:   res.NextKey,
			Limit: 4,
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, !res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(res.Items, 8, 9)

	// let's go backwards
	res, err = Paginate(table, store, &PaginationRequest{
		PageRequest: &query.PageRequest{
			Limit:      2,
			CountTotal: true,
			Reverse:    true,
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(res.Items, 9, 8)

	// a bit more
	res, err = Paginate(table, store, &PaginationRequest{
		PageRequest: &query.PageRequest{
			Key:     res.NextKey,
			Limit:   2,
			Reverse: true,
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(res.Items, 7, 6)

	// let's try an offset
	res, err = Paginate(table, store, &PaginationRequest{
		PageRequest: &query.PageRequest{
			Limit:      2,
			CountTotal: true,
			Offset:     3,
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 2, len(res.Cursors))
	assertGotItems(res.Items, 3, 4)

	// and reverse
	res, err = Paginate(table, store, &PaginationRequest{
		PageRequest: &query.PageRequest{
			Limit:      3,
			CountTotal: true,
			Offset:     5,
			Reverse:    true,
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)
	assert.Assert(t, res.NextKey != nil)
	assert.Equal(t, uint64(10), res.Total)
	assert.Assert(t, res.HaveMore)
	assert.Equal(t, 3, len(res.Cursors))
	assertGotItems(res.Items, 4, 3, 2)

	// now let's update some things
	for i := 0; i < 5; i++ {
		data[i].U64 = data[i].U64 * 2
		err = table.Save(store, data[i], SAVE_MODE_UPDATE)
		assert.NilError(t, err)
	}
	it, err = table.PrefixIterator(store, nil, IteratorOptions{})
	assert.NilError(t, err)
	// we should still get everything in the same order
	assertIteratorItems(it, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

	// let's use SAVE_MODE_DEFAULT and add something
	data = append(data, &testpb.A{U32: 9})
	err = table.Save(store, data[10], SAVE_MODE_DEFAULT)
	assert.NilError(t, err)
	found, err = table.Get(store, testutil.ValuesOf(uint32(9), int64(0), ""), &a)
	assert.NilError(t, err)
	assert.Assert(t, found)
	assert.DeepEqual(t, data[10], &a, protocmp.Transform())
	// and update it
	data[10].B = true
	err = table.Save(store, data[10], SAVE_MODE_DEFAULT)
	found, err = table.Get(store, testutil.ValuesOf(uint32(9), int64(0), ""), &a)
	assert.NilError(t, err)
	assert.Assert(t, found)
	assert.DeepEqual(t, data[10], &a, protocmp.Transform())
	// and iterate
	it, err = table.PrefixIterator(store, nil, IteratorOptions{})
	assert.NilError(t, err)
	assertIteratorItems(it, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// let's delete item 5
	key5 := testutil.ValuesOf(uint32(7), int64(-2), "abe")
	err = table.Delete(store, key5)
	assert.NilError(t, err)
	// it should be gone
	found, err = table.Has(store, key5)
	assert.NilError(t, err)
	assert.Assert(t, !found)
	// and missing from the iterator
	it, err = table.PrefixIterator(store, nil, IteratorOptions{})
	assert.NilError(t, err)
	assertIteratorItems(it, 0, 1, 2, 3, 4, 6, 7, 8, 9, 10)
}

func TestJSONExportImport(t *testing.T) {
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

		msg1, err := it.GetMessage()
		assert.NilError(t, err)
		msg2, err := it.GetMessage()
		assert.NilError(t, err)

		assert.DeepEqual(t, msg1, msg2, protocmp.Transform())
	}
}

func TestRandomTableData(t *testing.T) {
	testTable(t, TableDataGen(testutil.GenA, 100).Example().(*TableData))
}

func testTable(t *testing.T, tableData *TableData) {
	for _, index := range tableData.table.Indexes() {
		indexModel := &IndexModel{
			TableData: tableData,
			index:     index,
		}
		sort.Sort(indexModel)
		if _, ok := index.(UniqueIndex); ok {
			testUniqueIndex(t, indexModel)
		}
		testIndex(t, indexModel)
	}
}

func testUniqueIndex(t *testing.T, model *IndexModel) {
	index := model.index.(UniqueIndex)
	t.Logf("testing unique index %T %s", index, index.GetFieldNames())
	for i := 0; i < len(model.data); i++ {
		x := model.data[i]
		ks, _, err := index.(ormkv.IndexCodec).EncodeKeyFromMessage(x.ProtoReflect())
		assert.NilError(t, err)

		found, err := index.Has(model.store, ks)
		assert.NilError(t, err)
		assert.Assert(t, found)

		msg := model.table.MessageType().New().Interface()
		found, err = index.Get(model.store, ks, msg)
		assert.NilError(t, err)
		assert.Assert(t, found)
		assert.DeepEqual(t, x, msg, protocmp.Transform())
	}
}

func testIndex(t *testing.T, model *IndexModel) {
	index := model.index
	if index.IsFullyOrdered() {
		t.Logf("testing index %T %s", index, index.GetFieldNames())
		messageType := model.table.MessageType()

		it, err := model.index.PrefixIterator(model.store, nil, IteratorOptions{})
		assert.NilError(t, err)
		checkIteratorAgainstSlice(t, it, model.data, messageType)

		it, err = model.index.PrefixIterator(model.store, nil, IteratorOptions{Reverse: true})
		assert.NilError(t, err)
		checkIteratorAgainstSlice(t, it, reverseData(model.data), messageType)

		rapid.Check(t, func(t *rapid.T) {
			i := rapid.IntRange(0, len(model.data)-2).Draw(t, "i").(int)
			j := rapid.IntRange(i+1, len(model.data)-1).Draw(t, "j").(int)

			t.Logf("%d %d", i, j)
			start, _, err := model.index.(ormkv.IndexCodec).EncodeKeyFromMessage(model.data[i].ProtoReflect())
			assert.NilError(t, err)
			end, _, err := model.index.(ormkv.IndexCodec).EncodeKeyFromMessage(model.data[j].ProtoReflect())
			assert.NilError(t, err)
			it, err = model.index.RangeIterator(model.store, start, end, IteratorOptions{})
			assert.NilError(t, err)
			checkIteratorAgainstSlice(t, it, model.data[i:j], messageType)
		})
		//
		//it, err = model.index.RangeIterator(model.store, nil, nil, IteratorOptions{Reverse: true})
		//assert.NilError(t, err)
		//checkIteratorAgainstSlice(t, it, reverseData(model.data), messageType)
	} else {
		t.Logf("can't automatically test unordered index %T %s, TODO: test presence of all keys", index, index.GetFieldNames())
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

func checkIteratorAgainstSlice(t assert.TestingT, iterator Iterator, data []proto.Message, typ protoreflect.MessageType) {
	i := 0
	for iterator.Next() {
		msg := typ.New().Interface()
		err := iterator.UnmarshalMessage(msg)
		assert.NilError(t, err)
		assert.DeepEqual(t, data[i], msg, protocmp.Transform())
		i++
	}
}

func TableDataGen(elemGen *rapid.Generator, n int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *TableData {
		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)
		message := elemGen.Draw(t, "message").(proto.Message)
		table, err := Build(TableOptions{
			Prefix:      prefix,
			MessageType: message.ProtoReflect().Type(),
		})
		if err != nil {
			panic(err)
		}

		data := make([]proto.Message, n)
		store := memkv.NewMemIndexCommitmentStore()

		for i := 0; i < n; i++ {
			var err error
			for {
				message = elemGen.Draw(t, fmt.Sprintf("message[%d]", i)).(proto.Message)
				err = table.Save(store, message, SAVE_MODE_INSERT)
				if err == nil {
					break
				}
				if !sdkerrors.IsOf(err, ormerrors.PrimaryKeyConstraintViolation, ormerrors.UniqueKeyViolation) {
					panic(err)
				}
			}
			data[i] = message
		}

		return &TableData{
			data:  data,
			table: table,
			store: store,
		}
	})
}

type TableData struct {
	table Table
	data  []proto.Message
	store *memkv.IndexCommitmentStore
}

type IndexModel struct {
	*TableData
	index Index
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
