package ormtable_test

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"

	"google.golang.org/protobuf/testing/protocmp"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
)

func TestSingleton(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleSingleton{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)
	ctx := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())

	store, err := testpb.NewExampleSingletonTable(table)
	assert.NilError(t, err)

	val, err := store.Get(ctx)
	assert.NilError(t, err)
	assert.Assert(t, val != nil) // singletons are always set
	assert.NilError(t, store.Save(ctx, &testpb.ExampleSingleton{}))

	val.Foo = "abc"
	val.Bar = 3
	assert.NilError(t, store.Save(ctx, val))

	val2, err := store.Get(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, val2, protocmp.Transform())

	buf := &bytes.Buffer{}
	assert.NilError(t, table.ExportJSON(ctx, buf))
	assert.NilError(t, table.ValidateJSON(bytes.NewReader(buf.Bytes())))
	store2 := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())
	assert.NilError(t, table.ImportJSON(store2, bytes.NewReader(buf.Bytes())))

	val3, err := store.Get(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, val3, protocmp.Transform())
}
