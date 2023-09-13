package ormtable_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"cosmossdk.io/orm/internal/testkv"
	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/model/ormtable"
)

func TestAutoIncrementScenario(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleAutoIncrementTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	autoTable, ok := table.(ormtable.AutoIncrementTable)
	assert.Assert(t, ok)

	// first run tests with a split index-commitment store
	runAutoIncrementScenario(t, autoTable, ormtable.WrapContextDefault(testkv.NewSplitMemBackend()))

	// now run with shared store and debugging
	debugBuf := &strings.Builder{}
	store := testkv.NewDebugBackend(
		testkv.NewSharedMemBackend(),
		&testkv.EntryCodecDebugger{
			EntryCodec: table,
			Print:      func(s string) { debugBuf.WriteString(s + "\n") },
		},
	)
	runAutoIncrementScenario(t, autoTable, ormtable.WrapContextDefault(store))

	golden.Assert(t, debugBuf.String(), "test_auto_inc.golden")
	checkEncodeDecodeEntries(t, table, store.IndexStoreReader())
}

func runAutoIncrementScenario(t *testing.T, table ormtable.AutoIncrementTable, ctx context.Context) {
	t.Helper()
	store, err := testpb.NewExampleAutoIncrementTableTable(table)
	assert.NilError(t, err)

	err = store.Save(ctx, &testpb.ExampleAutoIncrementTable{Id: 5})
	assert.ErrorContains(t, err, "not found")

	ex1 := &testpb.ExampleAutoIncrementTable{X: "foo", Y: 5}
	assert.NilError(t, store.Save(ctx, ex1))
	assert.Equal(t, uint64(1), ex1.Id)
	curSeq, err := table.LastInsertedSequence(ctx)
	assert.NilError(t, err)
	assert.Equal(t, curSeq, uint64(1))

	ex2 := &testpb.ExampleAutoIncrementTable{X: "bar", Y: 10}
	newID, err := table.InsertReturningPKey(ctx, ex2)
	assert.NilError(t, err)
	assert.Equal(t, uint64(2), ex2.Id)
	assert.Equal(t, newID, ex2.Id)
	curSeq, err = table.LastInsertedSequence(ctx)
	assert.NilError(t, err)
	assert.Equal(t, curSeq, uint64(2))

	buf := &bytes.Buffer{}
	assert.NilError(t, table.ExportJSON(ctx, buf))
	assert.NilError(t, table.ValidateJSON(bytes.NewReader(buf.Bytes())))
	store2 := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	assert.NilError(t, table.ImportJSON(store2, bytes.NewReader(buf.Bytes())))
	assertTablesEqual(t, table, ctx, store2)

	// test edge case where we have deleted all entities but we're still exporting the sequence number
	assert.NilError(t, table.Delete(ctx, ex1))
	assert.NilError(t, table.Delete(ctx, ex2))
	buf = &bytes.Buffer{}
	assert.NilError(t, table.ExportJSON(ctx, buf))
	assert.NilError(t, table.ValidateJSON(bytes.NewReader(buf.Bytes())))
	golden.Assert(t, buf.String(), "trivial_auto_inc_export.golden")
	store3 := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	assert.NilError(t, table.ImportJSON(store3, bytes.NewReader(buf.Bytes())))
	ex1.Id = 0
	assert.NilError(t, table.Insert(store3, ex1))
	assert.Equal(t, uint64(3), ex1.Id) // should equal 3 because the sequence number 2 should have been imported from JSON
	curSeq, err = table.LastInsertedSequence(store3)
	assert.NilError(t, err)
	assert.Equal(t, curSeq, uint64(3))
}

func TestBadJSON(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleAutoIncrementTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	store := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	f, err := os.Open("testdata/bad_auto_inc.json")
	assert.NilError(t, err)
	assert.ErrorContains(t, table.ImportJSON(store, f), "invalid auto increment primary key")

	f, err = os.Open("testdata/bad_auto_inc2.json")
	assert.NilError(t, err)
	assert.ErrorContains(t, table.ImportJSON(store, f), "invalid auto increment primary key")
}
