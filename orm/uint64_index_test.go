package orm_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestUInt64Index(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")

	const anyPrefix = 0x10
	tableBuilder := orm.NewPrimaryKeyTableBuilder(anyPrefix, storeKey, &testdata.GroupMember{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	myIndex := orm.NewUInt64Index(tableBuilder, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]uint64, error) {
		return []uint64{uint64(val.(*testdata.GroupMember).Member[0])}, nil
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

	indexedKey := uint64('m')

	// Has
	assert.True(t, myIndex.Has(ctx, indexedKey))

	// Get
	it, err := myIndex.Get(ctx, indexedKey)
	require.NoError(t, err)
	var loaded testdata.GroupMember
	rowID, err := it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, uint64(1), orm.DecodeSequence(rowID))
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
			pageReq: &query.PageRequest{Key: []byte{byte('m')}},
			expErr:  true,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			it, err := myIndex.GetPaginated(ctx, indexedKey, tc.pageReq)
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
	it, err = myIndex.PrefixScan(ctx, 0, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, uint64(1), orm.DecodeSequence(rowID))
	require.Equal(t, m, loaded)

	// PrefixScan no match
	it, err = myIndex.PrefixScan(ctx, indexedKey+1, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.Error(t, orm.ErrIteratorDone, err)

	// ReversePrefixScan match
	it, err = myIndex.ReversePrefixScan(ctx, 0, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, uint64(1), orm.DecodeSequence(rowID))
	require.Equal(t, m, loaded)

	// ReversePrefixScan no match
	it, err = myIndex.ReversePrefixScan(ctx, indexedKey+1, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.Error(t, orm.ErrIteratorDone, err)
}

func TestUInt64MultiKeyAdapter(t *testing.T) {
	specs := map[string]struct {
		srcFunc orm.UInt64IndexerFunc
		exp     []orm.RowID
		expErr  error
	}{
		"single key": {
			srcFunc: func(value interface{}) ([]uint64, error) {
				return []uint64{1}, nil
			},
			exp: []orm.RowID{{0, 0, 0, 0, 0, 0, 0, 1}},
		},
		"multi key": {
			srcFunc: func(value interface{}) ([]uint64, error) {
				return []uint64{1, 1 << 56}, nil
			},
			exp: []orm.RowID{{0, 0, 0, 0, 0, 0, 0, 1}, {1, 0, 0, 0, 0, 0, 0, 0}},
		},
		"empty key": {
			srcFunc: func(value interface{}) ([]uint64, error) {
				return []uint64{}, nil
			},
			exp: []orm.RowID{},
		},
		"nil key": {
			srcFunc: func(value interface{}) ([]uint64, error) {
				return nil, nil
			},
			exp: []orm.RowID{},
		},
		"error case": {
			srcFunc: func(value interface{}) ([]uint64, error) {
				return nil, errors.New("test")
			},
			expErr: errors.New("test"),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			fn := orm.UInt64MultiKeyAdapter(spec.srcFunc)
			r, err := fn(nil)
			if spec.expErr != nil {
				require.Equal(t, spec.expErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, spec.exp, r)
		})
	}
}
