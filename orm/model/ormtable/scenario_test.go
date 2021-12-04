package ormtable

import (
	"bytes"
	"os"
	"testing"

	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/memkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
)

func TestScenario(t *testing.T) {
	//table, err := Build(TableOptions{
	//	MessageType: (&testpb.A{}).ProtoReflect().Type(),
	//})
	//assert.NilError(t, err)
	//store := memkv.NewMemIndexCommitmentStore()
	//
	//f, err := os.OpenFile("test_data.json", os.O_RDONLY, 0644)
	//assert.NilError(t, err)
	//
	//table.RangeIterator()
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
		have, err := it.Next()
		assert.NilError(t, err)

		have2, err := it2.Next()
		assert.NilError(t, err)

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
