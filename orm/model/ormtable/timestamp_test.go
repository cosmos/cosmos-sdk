package ormtable_test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/internal/testkv"
	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/model/ormtable"
)

func TestTimestampIndex(t *testing.T) {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.ExampleTimestamp{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)
	backend := testkv.NewDebugBackend(testkv.NewSplitMemBackend(), &testkv.EntryCodecDebugger{
		EntryCodec: table,
	})
	ctx := ormtable.WrapContextDefault(backend)
	store, err := testpb.NewExampleTimestampTable(table)
	assert.NilError(t, err)

	past, err := time.Parse("2006-01-02", "2000-01-01")
	assert.NilError(t, err)
	middle, err := time.Parse("2006-01-02", "2020-01-01")
	assert.NilError(t, err)
	future, err := time.Parse("2006-01-02", "2049-01-01")
	assert.NilError(t, err)

	pastPb, middlePb, futurePb := timestamppb.New(past), timestamppb.New(middle), timestamppb.New(future)
	timeOrder := []*timestamppb.Timestamp{pastPb, middlePb, futurePb}

	assert.NilError(t, store.Insert(ctx, &testpb.ExampleTimestamp{
		Name: "foo",
		Ts:   pastPb,
	}))
	assert.NilError(t, store.Insert(ctx, &testpb.ExampleTimestamp{
		Name: "bar",
		Ts:   middlePb,
	}))
	assert.NilError(t, store.Insert(ctx, &testpb.ExampleTimestamp{
		Name: "baz",
		Ts:   futurePb,
	}))

	from, to := testpb.ExampleTimestampTsIndexKey{}.WithTs(timestamppb.New(past)), testpb.ExampleTimestampTsIndexKey{}.WithTs(timestamppb.New(future))
	it, err := store.ListRange(ctx, from, to)
	assert.NilError(t, err)

	i := 0
	for it.Next() {
		v, err := it.Value()
		assert.NilError(t, err)
		assert.Equal(t, timeOrder[i].String(), v.Ts.String())
		i++
	}

	// insert a nil entry
	id, err := store.InsertReturningId(ctx, &testpb.ExampleTimestamp{
		Name: "nil",
		Ts:   nil,
	})
	assert.NilError(t, err)

	res, err := store.Get(ctx, id)
	assert.NilError(t, err)
	assert.Assert(t, res.Ts == nil)

	it, err = store.List(ctx, testpb.ExampleTimestampTsIndexKey{})
	assert.NilError(t, err)

	// make sure nils are ordered last
	timeOrder = append(timeOrder, nil)
	i = 0
	for it.Next() {
		v, err := it.Value()
		assert.NilError(t, err)
		assert.Assert(t, v != nil)
		x := timeOrder[i]
		if x == nil {
			assert.Assert(t, v.Ts == nil)
		} else {
			assert.Equal(t, x.String(), v.Ts.String())
		}
		i++
	}
	it.Close()

	// try iterating over just nil timestamps
	it, err = store.List(ctx, testpb.ExampleTimestampTsIndexKey{}.WithTs(nil))
	assert.NilError(t, err)
	assert.Assert(t, it.Next())
	res, err = it.Value()
	assert.NilError(t, err)
	assert.Assert(t, res.Ts == nil)
	assert.Assert(t, !it.Next())
	it.Close()
}
