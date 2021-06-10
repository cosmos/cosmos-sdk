package orm_test

import (
	"math"
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

func TestAutoUInt64PrefixScan(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	const (
		testTablePrefix = iota
		testTableSeqPrefix
	)
	tb := orm.NewAutoUInt64TableBuilder(testTablePrefix, testTableSeqPrefix, storeKey, &testdata.GroupInfo{}, cdc).Build()
	ctx := orm.NewMockContext()

	g1 := testdata.GroupInfo{
		Description: "my test 1",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}
	g2 := testdata.GroupInfo{
		Description: "my test 2",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}
	g3 := testdata.GroupInfo{
		Description: "my test 3",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}
	for _, g := range []testdata.GroupInfo{g1, g2, g3} {
		_, err := tb.Create(ctx, &g)
		require.NoError(t, err)
	}

	specs := map[string]struct {
		start, end uint64
		expResult  []testdata.GroupInfo
		expRowIDs  []orm.RowID
		expError   *errors.Error
		method     func(ctx orm.HasKVStore, start uint64, end uint64) (orm.Iterator, error)
	}{
		"first element": {
			start:     1,
			end:       2,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"first 2 elements": {
			start:     1,
			end:       3,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupInfo{g1, g2},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1), orm.EncodeSequence(2)},
		},
		"first 3 elements": {
			start:     1,
			end:       4,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupInfo{g1, g2, g3},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1), orm.EncodeSequence(2), orm.EncodeSequence(3)},
		},
		"search with max end": {
			start:     1,
			end:       math.MaxUint64,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupInfo{g1, g2, g3},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1), orm.EncodeSequence(2), orm.EncodeSequence(3)},
		},
		"2 to end": {
			start:     2,
			end:       5,
			method:    tb.PrefixScan,
			expResult: []testdata.GroupInfo{g2, g3},
			expRowIDs: []orm.RowID{orm.EncodeSequence(2), orm.EncodeSequence(3)},
		},
		"start before end should fail": {
			start:    2,
			end:      1,
			method:   tb.PrefixScan,
			expError: orm.ErrArgument,
		},
		"start equals end should fail": {
			start:    1,
			end:      1,
			method:   tb.PrefixScan,
			expError: orm.ErrArgument,
		},
		"reverse first element": {
			start:     1,
			end:       2,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(1)},
		},
		"reverse first 2 elements": {
			start:     1,
			end:       3,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g2, g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(2), orm.EncodeSequence(1)},
		},
		"reverse first 3 elements": {
			start:     1,
			end:       4,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g3, g2, g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(3), orm.EncodeSequence(2), orm.EncodeSequence(1)},
		},
		"reverse search with max end": {
			start:     1,
			end:       math.MaxUint64,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g3, g2, g1},
			expRowIDs: []orm.RowID{orm.EncodeSequence(3), orm.EncodeSequence(2), orm.EncodeSequence(1)},
		},
		"reverse 2 to end": {
			start:     2,
			end:       5,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.GroupInfo{g3, g2},
			expRowIDs: []orm.RowID{orm.EncodeSequence(3), orm.EncodeSequence(2)},
		},
		"reverse start before end should fail": {
			start:    2,
			end:      1,
			method:   tb.ReversePrefixScan,
			expError: orm.ErrArgument,
		},
		"reverse start equals end should fail": {
			start:    1,
			end:      1,
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
			var loaded []testdata.GroupInfo
			rowIDs, err := orm.ReadAll(it, &loaded)
			require.NoError(t, err)
			assert.Equal(t, spec.expResult, loaded)
			assert.Equal(t, spec.expRowIDs, rowIDs)
		})
	}
}
