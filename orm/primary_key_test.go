package orm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

func TestPrimaryKeyTablePrefixScan(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	const (
		testTablePrefix = iota
	)

	tb := orm.NewPrimaryKeyTableBuilder(testTablePrefix, storeKey, &testdata.GroupMember{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc).
		Build()

	ctx := orm.NewMockContext()

	const anyWeight = 1
	m1 := testdata.GroupMember{
		Group:  []byte("group-a"),
		Member: []byte("member-one"),
		Weight: anyWeight,
	}
	m2 := testdata.GroupMember{
		Group:  []byte("group-a"),
		Member: []byte("member-two"),
		Weight: anyWeight,
	}
	m3 := testdata.GroupMember{
		Group:  []byte("group-b"),
		Member: []byte("member-two"),
		Weight: anyWeight,
	}
	for _, g := range []testdata.GroupMember{m1, m2, m3} {
		require.NoError(t, tb.Create(ctx, &g))
	}

	specs := map[string]struct {
		start, end []byte
		expResult  []testdata.GroupMember
		expRowIDs  []orm.RowID
		expError   *errors.Error
		method     func(ctx orm.HasKVStore, start, end []byte) (orm.Iterator, error)
	}{
		"exact match with a single result": {
			start:     []byte("group-amember-one"), // == m1.PrimaryKey()
			end:       []byte("group-amember-two"), // == m2.PrimaryKey()
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m1},
			expRowIDs: []orm.RowID{m1.PrimaryKey()},
		},
		"one result by prefix": {
			start:     []byte("group-a"),
			end:       []byte("group-amember-two"), // == m2.PrimaryKey()
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m1},
			expRowIDs: []orm.RowID{m1.PrimaryKey()},
		},
		"multi key elements by group prefix": {
			start:     []byte("group-a"),
			end:       []byte("group-b"),
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m1, m2},
			expRowIDs: []orm.RowID{m1.PrimaryKey(), m2.PrimaryKey()},
		},
		"open end query with second group": {
			start:     []byte("group-b"),
			end:       nil,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m3},
			expRowIDs: []orm.RowID{m3.PrimaryKey()},
		},
		"open end query with all": {
			start:     []byte("group-a"),
			end:       nil,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m1, m2, m3},
			expRowIDs: []orm.RowID{m1.PrimaryKey(), m2.PrimaryKey(), m3.PrimaryKey()},
		},
		"open start query": {
			start:     nil,
			end:       []byte("group-b"),
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m1, m2},
			expRowIDs: []orm.RowID{m1.PrimaryKey(), m2.PrimaryKey()},
		},
		"open start and end query": {
			start:     nil,
			end:       nil,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m1, m2, m3},
			expRowIDs: []orm.RowID{m1.PrimaryKey(), m2.PrimaryKey(), m3.PrimaryKey()},
		},
		"all matching prefix": {
			start:     []byte("group"),
			end:       nil,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{m1, m2, m3},
			expRowIDs: []orm.RowID{m1.PrimaryKey(), m2.PrimaryKey(), m3.PrimaryKey()},
		},
		"non matching prefix": {
			start:     []byte("nobody"),
			end:       nil,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupMember{},
		},
		"start equals end": {
			start:    []byte("any"),
			end:      []byte("any"),
			method:   tb.PrefixScan,
			expError: orm.ErrArgument,
		},
		"start after end": {
			start:    []byte("b"),
			end:      []byte("a"),
			method:   tb.PrefixScan,
			expError: orm.ErrArgument,
		},
		"reverse: exact match with a single result": {
			start:     []byte("group-amember-one"), // == m1.PrimaryKey()
			end:       []byte("group-amember-two"), // == m2.PrimaryKey()
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m1},
			expRowIDs: []orm.RowID{m1.PrimaryKey()},
		},
		"reverse: one result by prefix": {
			start:     []byte("group-a"),
			end:       []byte("group-amember-two"), // == m2.PrimaryKey()
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m1},
			expRowIDs: []orm.RowID{m1.PrimaryKey()},
		},
		"reverse: multi key elements by group prefix": {
			start:     []byte("group-a"),
			end:       []byte("group-b"),
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m2, m1},
			expRowIDs: []orm.RowID{m2.PrimaryKey(), m1.PrimaryKey()},
		},
		"reverse: open end query with second group": {
			start:     []byte("group-b"),
			end:       nil,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m3},
			expRowIDs: []orm.RowID{m3.PrimaryKey()},
		},
		"reverse: open end query with all": {
			start:     []byte("group-a"),
			end:       nil,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m3, m2, m1},
			expRowIDs: []orm.RowID{m3.PrimaryKey(), m2.PrimaryKey(), m1.PrimaryKey()},
		},
		"reverse: open start query": {
			start:     nil,
			end:       []byte("group-b"),
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m2, m1},
			expRowIDs: []orm.RowID{m2.PrimaryKey(), m1.PrimaryKey()},
		},
		"reverse: open start and end query": {
			start:     nil,
			end:       nil,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m3, m2, m1},
			expRowIDs: []orm.RowID{m3.PrimaryKey(), m2.PrimaryKey(), m1.PrimaryKey()},
		},
		"reverse: all matching prefix": {
			start:     []byte("group"),
			end:       nil,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{m3, m2, m1},
			expRowIDs: []orm.RowID{m3.PrimaryKey(), m2.PrimaryKey(), m1.PrimaryKey()},
		},
		"reverse: non matching prefix": {
			start:     []byte("nobody"),
			end:       nil,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupMember{},
		},
		"reverse: start equals end": {
			start:    []byte("any"),
			end:      []byte("any"),
			method:   tb.ReversePrefixScan,
			expError: orm.ErrArgument,
		},
		"reverse: start after end": {
			start:    []byte("b"),
			end:      []byte("a"),
			method:   tb.ReversePrefixScan,
			expError: orm.ErrArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			it, err := spec.method(ctx, spec.start, spec.end)
			require.True(t, spec.expError.Is(err), "expected #+v but got #+v", spec.expError, err)
			if spec.expError != nil {
				return
			}
			var loaded []testdata.GroupMember
			rowIDs, err := orm.ReadAll(it, &loaded)
			require.NoError(t, err)
			assert.Equal(t, spec.expResult, loaded)
			assert.Equal(t, spec.expRowIDs, rowIDs)
		})
	}
}

func TestContains(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	const testTablePrefix = iota

	tb := orm.NewPrimaryKeyTableBuilder(testTablePrefix, storeKey, &testdata.GroupMember{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc).
		Build()

	ctx := orm.NewMockContext()

	myPersistentObj := testdata.GroupMember{
		Group:  []byte("group-a"),
		Member: []byte("member-one"),
		Weight: 1,
	}
	err := tb.Create(ctx, &myPersistentObj)
	require.NoError(t, err)

	specs := map[string]struct {
		src orm.PrimaryKeyed
		exp bool
	}{

		"same object": {src: &myPersistentObj, exp: true},
		"clone": {
			src: &testdata.GroupMember{
				Group:  []byte("group-a"),
				Member: []byte("member-one"),
				Weight: 1,
			},
			exp: true,
		},
		"different primary key": {
			src: &testdata.GroupMember{
				Group:  []byte("another group"),
				Member: []byte("member-one"),
				Weight: 1,
			},
			exp: false,
		},
		"different type, same key": {
			src: mockPrimaryKeyed{&myPersistentObj},
			exp: false,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got := tb.Contains(ctx, spec.src)
			assert.Equal(t, spec.exp, got)
		})
	}
}

type mockPrimaryKeyed struct {
	*testdata.GroupMember
}
