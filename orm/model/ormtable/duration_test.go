package ormtable_test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/internal/testkv"
	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/model/ormtable"
)

func TestDurationIndex(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleDuration{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)
	backend := testkv.NewDebugBackend(testkv.NewSplitMemBackend(), &testkv.EntryCodecDebugger{
		EntryCodec: table,
	})
	ctx := ormtable.WrapContextDefault(backend)
	store, err := testpb.NewExampleDurationTable(table)
	assert.NilError(t, err)

	neg, err := time.ParseDuration("-1h")
	assert.NilError(t, err)
	zero, err := time.ParseDuration("0")
	assert.NilError(t, err)
	pos, err := time.ParseDuration("11000ms")
	assert.NilError(t, err)

	negPb, zeroPb, posPb := durationpb.New(neg), durationpb.New(zero), durationpb.New(pos)
	durOrder := []*durationpb.Duration{negPb, zeroPb, posPb}

	assert.NilError(t, store.Insert(ctx, &testpb.ExampleDuration{
		Name: "foo",
		Dur:  negPb,
	}))
	assert.NilError(t, store.Insert(ctx, &testpb.ExampleDuration{
		Name: "bar",
		Dur:  zeroPb,
	}))
	assert.NilError(t, store.Insert(ctx, &testpb.ExampleDuration{
		Name: "baz",
		Dur:  posPb,
	}))

	from, to := testpb.ExampleDurationDurIndexKey{}.WithDur(durationpb.New(neg)),
		testpb.ExampleDurationDurIndexKey{}.WithDur(durationpb.New(pos))
	it, err := store.ListRange(ctx, from, to)
	assert.NilError(t, err)

	i := 0
	for it.Next() {
		v, err := it.Value()
		assert.NilError(t, err)
		assert.Equal(t, durOrder[i].String(), v.Dur.String())
		i++
	}

	// insert a nil entry
	id, err := store.InsertReturningId(ctx, &testpb.ExampleDuration{
		Name: "nil",
		Dur:  nil,
	})
	assert.NilError(t, err)

	res, err := store.Get(ctx, id)
	assert.NilError(t, err)
	assert.Assert(t, res.Dur == nil)

	it, err = store.List(ctx, testpb.ExampleDurationDurIndexKey{})
	assert.NilError(t, err)

	// make sure nils are ordered last
	durOrder = append(durOrder, nil)
	i = 0
	for it.Next() {
		v, err := it.Value()
		assert.NilError(t, err)
		assert.Assert(t, v != nil)
		x := durOrder[i]
		if x == nil {
			assert.Assert(t, v.Dur == nil)
		} else {
			assert.Equal(t, x.String(), v.Dur.String())
		}
		i++
	}
	it.Close()

	// try iterating over just nil timestamps
	it, err = store.List(ctx, testpb.ExampleDurationDurIndexKey{}.WithDur(nil))
	assert.NilError(t, err)
	assert.Assert(t, it.Next())
	res, err = it.Value()
	assert.NilError(t, err)
	assert.Assert(t, res.Dur == nil)
	assert.Assert(t, !it.Next())
	it.Close()
}
