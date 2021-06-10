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
	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestIndexPrefixScan(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	storeKey := sdk.NewKVStoreKey("test")
	const (
		testTablePrefix = iota
		testTableSeqPrefix
	)
	tBuilder := orm.NewAutoUInt64TableBuilder(testTablePrefix, testTableSeqPrefix, storeKey, &testdata.GroupInfo{}, cdc)
	idx := orm.NewIndex(tBuilder, GroupByAdminIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		return []orm.RowID{[]byte(val.(*testdata.GroupInfo).Admin)}, nil
	})
	tb := tBuilder.Build()
	ctx := orm.NewMockContext()

	g1 := testdata.GroupInfo{
		Description: "my test 1",
		Admin:       sdk.AccAddress([]byte("admin-address-a")),
	}
	g2 := testdata.GroupInfo{
		Description: "my test 2",
		Admin:       sdk.AccAddress([]byte("admin-address-b")),
	}
	g3 := testdata.GroupInfo{
		Description: "my test 3",
		Admin:       sdk.AccAddress([]byte("admin-address-b")),
	}
	for _, g := range []testdata.GroupInfo{g1, g2, g3} {
		_, err := tb.Create(ctx, &g)
		require.NoError(t, err)
	}

	specs := map[string]struct {
		start, end []byte
		expResult  []testdata.GroupInfo
		expRowIDs  []orm.RowID
		expError   *errors.Error
		method     func(ctx orm.HasKVStore, start, end []byte) (orm.Iterator, error)
	}{
		"exact match with a single result": {
			start:     []byte("admin-address-a"),
			end:       []byte("admin-address-b"),
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"one result by prefix": {
			start:     []byte("admin-address"),
			end:       []byte("admin-address-b"),
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"multi key elements by exact match": {
			start:     []byte("admin-address-b"),
			end:       []byte("admin-address-c"),
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{g2, g3},
			expRowIDs: []orm.RowID{orm.EncodeSequence(2), orm.EncodeSequence(3)},
		},
		"open end query": {
			start:     []byte("admin-address-b"),
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{g2, g3},
			expRowIDs: []orm.RowID{orm.EncodeSequence(2), orm.EncodeSequence(3)},
		},
		"open start query": {
			start:     nil,
			end:       []byte("admin-address-b"),
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"open start and end query": {
			start:     nil,
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{g1, g2, g3},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1), orm.EncodeSequence(2), orm.EncodeSequence(3)},
		},
		"all matching prefix": {
			start:     []byte("admin"),
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{g1, g2, g3},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1), orm.EncodeSequence(2), orm.EncodeSequence(3)},
		},
		"non matching prefix": {
			start:     []byte("nobody"),
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []testdata.GroupInfo{},
		},
		"start equals end": {
			start:    []byte("any"),
			end:      []byte("any"),
			method:   idx.PrefixScan,
			expError: orm.ErrArgument,
		},
		"start after end": {
			start:    []byte("b"),
			end:      []byte("a"),
			method:   idx.PrefixScan,
			expError: orm.ErrArgument,
		},
		"reverse: exact match with a single result": {
			start:     []byte("admin-address-a"),
			end:       []byte("admin-address-b"),
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"reverse: one result by prefix": {
			start:     []byte("admin-address"),
			end:       []byte("admin-address-b"),
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"reverse: multi key elements by exact match": {
			start:     []byte("admin-address-b"),
			end:       []byte("admin-address-c"),
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g3, g2},
			expRowIDs: []orm.RowID{orm.EncodeSequence(3), orm.EncodeSequence(2)},
		},
		"reverse: open end query": {
			start:     []byte("admin-address-b"),
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g3, g2},
			expRowIDs: []orm.RowID{orm.EncodeSequence(3), orm.EncodeSequence(2)},
		},
		"reverse: open start query": {
			start:     nil,
			end:       []byte("admin-address-b"),
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"reverse: open start and end query": {
			start:     nil,
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g3, g2, g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(3), orm.EncodeSequence(2), orm.EncodeSequence(1)},
		},
		"reverse: all matching prefix": {
			start:     []byte("admin"),
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g3, g2, g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(3), orm.EncodeSequence(2), orm.EncodeSequence(1)},
		},
		"reverse: non matching prefix": {
			start:     []byte("nobody"),
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []testdata.GroupInfo{},
		},
		"reverse: start equals end": {
			start:    []byte("any"),
			end:      []byte("any"),
			method:   idx.ReversePrefixScan,
			expError: orm.ErrArgument,
		},
		"reverse: start after end": {
			start:    []byte("b"),
			end:      []byte("a"),
			method:   idx.ReversePrefixScan,
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
			var loaded []testdata.GroupInfo
			rowIDs, err := orm.ReadAll(it, &loaded)
			require.NoError(t, err)
			assert.Equal(t, spec.expResult, loaded)
			assert.Equal(t, spec.expRowIDs, rowIDs)
		})
	}
}

func TestUniqueIndex(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")

	tableBuilder := orm.NewPrimaryKeyTableBuilder(GroupMemberTablePrefix, storeKey, &testdata.GroupMember{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	uniqueIdx := orm.NewUniqueIndex(tableBuilder, 0x10, func(val interface{}) (orm.RowID, error) {
		return []byte{val.(*testdata.GroupMember).Member[0]}, nil
	})
	myTable := tableBuilder.Build()

	ctx := orm.NewMockContext()

	m := testdata.GroupMember{
		Group:  sdk.AccAddress(orm.EncodeSequence(1)),
		Member: sdk.AccAddress([]byte("member-address")),
		Weight: 10,
	}
	err := myTable.Create(ctx, &m)
	require.NoError(t, err)

	indexedKey := []byte{byte('m')}

	// Has
	assert.True(t, uniqueIdx.Has(ctx, indexedKey))

	// Get
	it, err := uniqueIdx.Get(ctx, indexedKey)
	require.NoError(t, err)
	var loaded testdata.GroupMember
	rowID, err := it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, orm.RowID(m.PrimaryKey()), rowID)
	require.Equal(t, m, loaded)

	// GetPaginated
	cases := map[string]struct {
		pageReq *query.PageRequest
		expErr  bool
	}{
		"nil key": {
			pageReq: &query.PageRequest{Key: nil},
			expErr:  false,
		},
		"after indexed key": {
			pageReq: &query.PageRequest{Key: indexedKey},
			expErr:  true,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			it, err := uniqueIdx.GetPaginated(ctx, indexedKey, tc.pageReq)
			require.NoError(t, err)
			rowID, err := it.LoadNext(&loaded)
			if tc.expErr { // iterator done
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, orm.RowID(m.PrimaryKey()), rowID)
				require.Equal(t, m, loaded)
			}
		})
	}

	// PrefixScan match
	it, err = uniqueIdx.PrefixScan(ctx, []byte{byte('m')}, []byte{byte('n')})
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, orm.RowID(m.PrimaryKey()), rowID)
	require.Equal(t, m, loaded)

	// PrefixScan no match
	it, err = uniqueIdx.PrefixScan(ctx, []byte{byte('n')}, nil)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.Error(t, orm.ErrIteratorDone, err)

	// ReversePrefixScan match
	it, err = uniqueIdx.ReversePrefixScan(ctx, []byte{byte('a')}, []byte{byte('z')})
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, orm.RowID(m.PrimaryKey()), rowID)
	require.Equal(t, m, loaded)

	// ReversePrefixScan no match
	it, err = uniqueIdx.ReversePrefixScan(ctx, []byte{byte('l')}, nil)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.Error(t, orm.ErrIteratorDone, err)
	// create with same index key should fail
	new := testdata.GroupMember{
		Group:  sdk.AccAddress(orm.EncodeSequence(1)),
		Member: sdk.AccAddress([]byte("my-other")),
		Weight: 10,
	}
	err = myTable.Create(ctx, &new)
	require.Error(t, orm.ErrUniqueConstraint, err)

	// and when delete
	err = myTable.Delete(ctx, &m)
	require.NoError(t, err)

	// then no persistent element
	assert.False(t, uniqueIdx.Has(ctx, indexedKey))
}

func TestPrefixRange(t *testing.T) {
	cases := map[string]struct {
		src      []byte
		expStart []byte
		expEnd   []byte
		expPanic bool
	}{
		"normal":                 {src: []byte{1, 3, 4}, expStart: []byte{1, 3, 4}, expEnd: []byte{1, 3, 5}},
		"normal short":           {src: []byte{79}, expStart: []byte{79}, expEnd: []byte{80}},
		"empty case":             {src: []byte{}},
		"roll-over example 1":    {src: []byte{17, 28, 255}, expStart: []byte{17, 28, 255}, expEnd: []byte{17, 29, 0}},
		"roll-over example 2":    {src: []byte{15, 42, 255, 255}, expStart: []byte{15, 42, 255, 255}, expEnd: []byte{15, 43, 0, 0}},
		"pathological roll-over": {src: []byte{255, 255, 255, 255}, expStart: []byte{255, 255, 255, 255}},
		"nil prohibited":         {expPanic: true},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if tc.expPanic {
				require.Panics(t, func() {
					orm.PrefixRange(tc.src)
				})
				return
			}
			start, end := orm.PrefixRange(tc.src)
			assert.Equal(t, tc.expStart, start)
			assert.Equal(t, tc.expEnd, end)
		})
	}
}
