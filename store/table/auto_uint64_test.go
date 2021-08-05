package table

import (
	"math"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestAutoUInt64PrefixScan(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	const (
		testTablePrefix = iota
		testTableSeqPrefix
	)
	tb := NewAutoUInt64TableBuilder(testTablePrefix, testTableSeqPrefix, &testdata.TableModel{}, cdc).Build()
	ctx := NewMockContext()
	store := ctx.KVStore(sdk.NewKVStoreKey("test"))

	m1 := testdata.TableModel{
		Id:   1,
		Name: "name 1",
	}
	m2 := testdata.TableModel{
		Id:   2,
		Name: "name 2",
	}
	m3 := testdata.TableModel{
		Id:   3,
		Name: "name 3",
	}

	for _, m := range []testdata.TableModel{m1, m2, m3} {
		_, err := tb.Create(store, &m)
		require.NoError(t, err)
	}

	specs := map[string]struct {
		start, end uint64
		expResult  []testdata.TableModel
		expRowIDs  []RowID
		expError   *errors.Error
		method     func(store sdk.KVStore, start uint64, end uint64) (Iterator, error)
	}{
		"first element": {
			start:     1,
			end:       2,
			method:    tb.PrefixScan,
			expResult: []testdata.TableModel{m1},
			expRowIDs: []RowID{EncodeSequence(1)},
		},
		"first 2 elements": {
			start:     1,
			end:       3,
			method:    tb.PrefixScan,
			expResult: []testdata.TableModel{m1, m2},
			expRowIDs: []RowID{EncodeSequence(1), EncodeSequence(2)},
		},
		"first 3 elements": {
			start:     1,
			end:       4,
			method:    tb.PrefixScan,
			expResult: []testdata.TableModel{m1, m2, m3},
			expRowIDs: []RowID{EncodeSequence(1), EncodeSequence(2), EncodeSequence(3)},
		},
		"search with max end": {
			start:     1,
			end:       math.MaxUint64,
			method:    tb.PrefixScan,
			expResult: []testdata.TableModel{m1, m2, m3},
			expRowIDs: []RowID{EncodeSequence(1), EncodeSequence(2), EncodeSequence(3)},
		},
		"2 to end": {
			start:     2,
			end:       5,
			method:    tb.PrefixScan,
			expResult: []testdata.TableModel{m2, m3},
			expRowIDs: []RowID{EncodeSequence(2), EncodeSequence(3)},
		},
		"start before end should fail": {
			start:    2,
			end:      1,
			method:   tb.PrefixScan,
			expError: ErrArgument,
		},
		"start equals end should fail": {
			start:    1,
			end:      1,
			method:   tb.PrefixScan,
			expError: ErrArgument,
		},
		"reverse first element": {
			start:     1,
			end:       2,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.TableModel{m1},
			expRowIDs: []RowID{EncodeSequence(1)},
		},
		"reverse first 2 elements": {
			start:     1,
			end:       3,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.TableModel{m2, m1},
			expRowIDs: []RowID{EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse first 3 elements": {
			start:     1,
			end:       4,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.TableModel{m3, m2, m1},
			expRowIDs: []RowID{EncodeSequence(3), EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse search with max end": {
			start:     1,
			end:       math.MaxUint64,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.TableModel{m3, m2, m1},
			expRowIDs: []RowID{EncodeSequence(3), EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse 2 to end": {
			start:     2,
			end:       5,
			method:    tb.ReversePrefixScan,
			expResult: []testdata.TableModel{m3, m2},
			expRowIDs: []RowID{EncodeSequence(3), EncodeSequence(2)},
		},
		"reverse start before end should fail": {
			start:    2,
			end:      1,
			method:   tb.ReversePrefixScan,
			expError: ErrArgument,
		},
		"reverse start equals end should fail": {
			start:    1,
			end:      1,
			method:   tb.ReversePrefixScan,
			expError: ErrArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			it, err := spec.method(store, spec.start, spec.end)
			require.True(t, spec.expError.Is(err), "expected #+v but got #+v", spec.expError, err)
			if spec.expError != nil {
				return
			}

			var loaded []testdata.TableModel
			rowIDs, err := ReadAll(store, it, &loaded)
			require.NoError(t, err)
			assert.Equal(t, spec.expResult, loaded)
			assert.Equal(t, spec.expRowIDs, rowIDs)
		})
	}
}
