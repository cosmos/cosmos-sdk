package ormtable

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"sort"
	"testing"

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

func TestTable(t *testing.T) {
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
	} else {
		t.Logf("can't automatically test unordered index %T %s", index, index.GetFieldNames())
	}

	//
	//i := rapid.IntRange(0, len(model.data)).Draw(t, "i").(int)
	//j := rapid.IntRange(i, len(model.data)).Draw(t, "j").(int)
	//
	//start, _, err := model.index.EncodeKeyFromMessage(model.data[i].ProtoReflect())
	//assert.NilError(t, err)
	//end, _, err := model.index.EncodeKeyFromMessage(model.data[j].ProtoReflect())
	//assert.NilError(t, err)
	//it, err = model.index.RangeIterator(model.store, start, end, ormindex.IteratorOptions{})
	//assert.NilError(t, err)
	//checkIteratorAgainstSlice(t, it, model.data[i:j], model.typ)

	//it, err = model.index.RangeIterator(model.store, nil, nil, IteratorOptions{Reverse: true})
	//assert.NilError(t, err)
	//checkIteratorAgainstSlice(t, it, reverseData(model.data), model.typ)
}

func reverseData(data []proto.Message) []proto.Message {
	n := len(data)
	reverse := make([]proto.Message, n)
	for i := 0; i < n; i++ {
		reverse[n-i-1] = data[i]
	}
	return reverse
}

func checkIteratorAgainstSlice(t *testing.T, iterator Iterator, data []proto.Message, typ protoreflect.MessageType) {
	i := 0
	for {
		have, err := iterator.Next()
		assert.NilError(t, err)
		if !have {
			break
		}

		msg := typ.New().Interface()
		err = iterator.UnmarshalMessage(msg)
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
				err = table.Save(store, message, SAVE_MODE_CREATE)
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
