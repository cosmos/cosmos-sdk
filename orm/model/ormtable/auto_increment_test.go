package ormtable_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

func TestAutoIncrementScenario(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleAutoIncrementTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	// first run tests with a split index-commitment store
	runAutoIncrementScenario(t, table, ormtable.WrapContextDefault(testkv.NewSplitMemBackend()))

	// now run with shared store and debugging
	debugBuf := &strings.Builder{}
	store := testkv.NewDebugBackend(
		testkv.NewSharedMemBackend(),
		&testkv.EntryCodecDebugger{
			EntryCodec: table,
			Print:      func(s string) { debugBuf.WriteString(s + "\n") },
		},
	)
	runAutoIncrementScenario(t, table, ormtable.WrapContextDefault(store))

	golden.Assert(t, debugBuf.String(), "test_auto_inc.golden")
	checkEncodeDecodeEntries(t, table, store.IndexStoreReader())
}

func runAutoIncrementScenario(t *testing.T, table ormtable.Table, context context.Context) {
	err := table.Save(context, &testpb.ExampleAutoIncrementTable{Id: 5})
	assert.ErrorContains(t, err, "update")

	ex1 := &testpb.ExampleAutoIncrementTable{X: "foo", Y: 5}
	assert.NilError(t, table.Save(context, ex1))
	assert.Equal(t, uint64(1), ex1.Id)

	buf := &bytes.Buffer{}
	assert.NilError(t, table.ExportJSON(context, buf))
	assert.NilError(t, table.ValidateJSON(bytes.NewReader(buf.Bytes())))
	store2 := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	assert.NilError(t, table.ImportJSON(store2, bytes.NewReader(buf.Bytes())))
	assertTablesEqual(t, table, context, store2)
}

func TestBadJSON(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleAutoIncrementTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	store := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	f, err := os.Open("testdata/bad_auto_inc.json")
	assert.NilError(t, err)
	assert.ErrorContains(t, table.ImportJSON(store, f), "invalid ID")

	f, err = os.Open("testdata/bad_auto_inc2.json")
	assert.NilError(t, err)
	assert.ErrorContains(t, table.ImportJSON(store, f), "invalid ID")
}
