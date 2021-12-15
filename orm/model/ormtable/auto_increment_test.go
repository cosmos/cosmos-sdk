package ormtable

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"gotest.tools/v3/golden"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
)

func TestAutoIncrementScenario(t *testing.T) {
	table, err := Build(Options{
		MessageType: (&testpb.ExampleAutoIncrementTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	// first run tests with a split index-commitment store
	runAutoIncrementScenario(t, table, testkv.NewSplitMemBackend())

	// now run with shared store and debugging
	debugBuf := &strings.Builder{}
	sharedStore := testkv.NewSharedMemBackend()
	store := testkv.NewDebugBackend(
		sharedStore,
		&testkv.EntryCodecDebugger{
			EntryCodec: table,
			Print:      func(s string) { debugBuf.WriteString(s + "\n") },
		},
	)
	runAutoIncrementScenario(t, table, store)

	golden.Assert(t, debugBuf.String(), "test_auto_inc.golden")
	checkEncodeDecodeEntries(t, table, store.IndexStoreReader())
}

func runAutoIncrementScenario(t *testing.T, table Table, store kvstore.Backend) {
	err := table.Save(store, &testpb.ExampleAutoIncrementTable{Id: 5}, SAVE_MODE_DEFAULT)
	assert.ErrorContains(t, err, "update")

	ex1 := &testpb.ExampleAutoIncrementTable{X: "foo", Y: 5}
	assert.NilError(t, table.Save(store, ex1, SAVE_MODE_DEFAULT))
	assert.Equal(t, uint64(1), ex1.Id)

	buf := &bytes.Buffer{}
	assert.NilError(t, table.ExportJSON(store, buf))
	golden.Assert(t, string(buf.Bytes()), "auto_inc_json.golden")

	assert.NilError(t, table.ValidateJSON(bytes.NewReader(buf.Bytes())))
	store2 := testkv.NewSplitMemBackend()
	assert.NilError(t, table.ImportJSON(store2, bytes.NewReader(buf.Bytes())))
	assertTablesEqual(t, table, store, store2)
}

func TestBadJSON(t *testing.T) {
	table, err := Build(Options{
		MessageType: (&testpb.ExampleAutoIncrementTable{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	store := testkv.NewSplitMemBackend()
	f, err := os.Open("testdata/bad_auto_inc.json")
	assert.NilError(t, err)
	assert.ErrorContains(t, table.ImportJSON(store, f), "invalid ID")

	f, err = os.Open("testdata/bad_auto_inc2.json")
	assert.NilError(t, err)
	assert.ErrorContains(t, table.ImportJSON(store, f), "invalid ID")
}
