package ormtable

import (
	"fmt"
	"sort"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/internal/memkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
	"github.com/cosmos/cosmos-sdk/orm/model/ormindex"
	"github.com/cosmos/cosmos-sdk/orm/model/ormiterator"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestErr(t *testing.T) {
	//assert.Assert(t, errors.Is(errors.Wrapf(ormerrors.PrimaryKeyConstraintViolation.Wrapf(), "test"), ormerrors.PrimaryKeyConstraintViolation))
}

func TestTable(t *testing.T) {
	tableData := TableDataGen(100).Example().(*TableData)
	for _, index := range tableData.table.Indexes() {
		if uniq, ok := index.(ormindex.UniqueIndex); ok {
			testUniqueIndex(t, tableData, uniq)
		}
	}
}

func TestUniqueKeyIndex(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		//primaryKeyModel := TableDataGen(100).Draw(t, "primaryKeyModel").(*IndexModel)
		//keyPrefix := AppendVarUInt32(primaryKeyModel.prefix, 1)
		//pkFields := primaryKeyModel.index.GetFieldNames()
		//// to ensure uniqueness we just reverse the primary key
		//n := len(pkFields)
		//uniqueFields := make([]protoreflect.Name, n)
		//for i := 0; i < n; i++ {
		//	uniqueFields[n-i-1] = pkFields[i]
		//}
		//
		//uniqueKeyCodec, err := ormkv.NewUniqueKeyCodec(
		//	keyPrefix,
		//	primaryKeyModel.typ.Descriptor(),
		//	uniqueFields,
		//	pkFields,
		//)
		//assert.NilError(t, err)
		//
		//index := ormindex.NewUniqueKeyIndex(uniqueKeyCodec, primaryKeyModel.index.(*ormindex.PrimaryKeyIndex))
		//
		//for _, datum := range primaryKeyModel.data {
		//	err = index.OnCreate(primaryKeyModel.store.IndexStore(), datum.ProtoReflect())
		//	assert.NilError(t, err)
		//}
		//
		//model := &IndexModel{
		//	typ:    primaryKeyModel.typ,
		//	prefix: primaryKeyModel.prefix,
		//	data:   primaryKeyModel.data,
		//	index:  index,
		//	store:  primaryKeyModel.store,
		//}
		//
		//sort.Sort(model)
		//
		//testUniqueIndex(t, model)
	})
}

func testUniqueIndex(t *testing.T, tableData *TableData, index ormindex.UniqueIndex) {
	model := &IndexModel{
		TableData: tableData,
		index:     index,
	}
	sort.Sort(model)
	for i := 0; i < len(model.data); i++ {
		x := model.data[i]
		ks, _, err := index.EncodeKeyFromMessage(x.ProtoReflect())
		assert.NilError(t, err)

		found, err := index.Has(model.store, ks)
		assert.NilError(t, err)
		assert.Assert(t, found)

		msg := tableData.table.MessageType().New().Interface()
		found, err = index.Get(model.store, ks, msg)
		assert.NilError(t, err)
		assert.Assert(t, found)
		assert.DeepEqual(t, x, msg, protocmp.Transform())
	}
}

//func testIndex(t *rapid.T, model *IndexModel) {
//	it, err := model.index.PrefixIterator(model.store, nil, ormindex.IteratorOptions{})
//	assert.NilError(t, err)
//	checkIteratorAgainstSlice(t, it, model.data, model.typ)
//
//	it, err = model.index.PrefixIterator(model.store, nil, ormindex.IteratorOptions{Reverse: true})
//	assert.NilError(t, err)
//	checkIteratorAgainstSlice(t, it, reverseData(model.data), model.typ)
//
//	i := rapid.IntRange(0, len(model.data)).Draw(t, "i").(int)
//	j := rapid.IntRange(i, len(model.data)).Draw(t, "j").(int)
//
//	start, _, err := model.index.EncodeKeyFromMessage(model.data[i].ProtoReflect())
//	assert.NilError(t, err)
//	end, _, err := model.index.EncodeKeyFromMessage(model.data[j].ProtoReflect())
//	assert.NilError(t, err)
//	it, err = model.index.RangeIterator(model.store, start, end, ormindex.IteratorOptions{})
//	assert.NilError(t, err)
//	checkIteratorAgainstSlice(t, it, model.data[i:j], model.typ)
//
//	//it, err = model.index.RangeIterator(model.store, nil, nil, IteratorOptions{Reverse: true})
//	//assert.NilError(t, err)
//	//checkIteratorAgainstSlice(t, it, reverseData(model.data), model.typ)
//}

func reverseData(data []proto.Message) []proto.Message {
	n := len(data)
	reverse := make([]proto.Message, n)
	for i := 0; i < n; i++ {
		reverse[n-i-1] = data[i]
	}
	return reverse
}

func checkIteratorAgainstSlice(t *rapid.T, iterator ormiterator.Iterator, data []proto.Message, typ protoreflect.MessageType) {
	i := 0
	for {
		have, err := iterator.Next()
		assert.NilError(t, err)
		if !have {
			break
		}

		msg := typ.New().Interface()
		err = iterator.GetMessage(msg)
		assert.NilError(t, err)
		assert.DeepEqual(t, data[i], msg, protocmp.Transform())
		i++
	}
}

var aType = (&testpb.A{}).ProtoReflect().Type()

func TableDataGen(n int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *TableData {
		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)
		table, err := Build(TableOptions{
			Prefix:      prefix,
			MessageType: aType,
		})
		if err != nil {
			panic(err)
		}
		//assert.NilError(t, err)

		data := make([]proto.Message, n)
		store := memkv.NewIndexCommitmentStore()

		for i := 0; i < n; i++ {
			var a proto.Message
			var err error
			for {
				a = testutil.GenA.Draw(t, fmt.Sprintf("a%d", i)).(*testpb.A)
				err = table.Save(store, a, SAVE_MODE_CREATE)
				if err == nil {
					break
				}
				if !sdkerrors.IsOf(err, ormerrors.PrimaryKeyConstraintViolation, ormerrors.UniqueKeyViolation) {
					panic(err)
				}
			}
			data[i] = a
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
	index ormindex.Index
}

func (m *IndexModel) Len() int {
	return len(m.data)
}

func (m *IndexModel) Less(i, j int) bool {
	is, _, err := m.index.EncodeKeyFromMessage(m.data[i].ProtoReflect())
	if err != nil {
		panic(err)
	}
	js, _, err := m.index.EncodeKeyFromMessage(m.data[j].ProtoReflect())
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
