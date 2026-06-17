package blockstm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/test-go/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

func TestEmptyMVData(t *testing.T) {
	ctx := context.Background()
	data := NewMVData(1)
	value, _, estimate := data.Read(ctx, []byte("a"), 1)
	require.False(t, estimate)
	require.Nil(t, value)
}

func KV(kv ...[]byte) *MemDB {
	db := NewMemDB()
	for i := 0; i < len(kv); i += 2 {
		db.OverlaySet(kv[i], kv[i+1])
	}
	return db
}

func TestMVData(t *testing.T) {
	ctx := context.Background()
	data := NewMVData(10)

	// read closest version
	data.Consolidate(ctx, TxnVersion{Index: 1, Incarnation: 1}, KV([]byte("a"), []byte("1")))
	data.Consolidate(ctx, TxnVersion{Index: 2, Incarnation: 1}, KV([]byte("a"), []byte("2")))
	data.Consolidate(ctx, TxnVersion{Index: 3, Incarnation: 1}, KV([]byte("a"), []byte("3")))
	data.Consolidate(ctx, TxnVersion{Index: 2, Incarnation: 1}, KV([]byte("a"), []byte("2"), []byte("b"), []byte("2")))

	// read closest version
	value, _, estimate := data.Read(ctx, []byte("a"), 1)
	require.False(t, estimate)
	require.Nil(t, value)

	// read closest version
	value, version, estimate := data.Read(ctx, []byte("a"), 4)
	require.False(t, estimate)
	require.Equal(t, []byte("3"), value)
	require.Equal(t, TxnVersion{Index: 3, Incarnation: 1}, version)

	// read closest version
	value, version, estimate = data.Read(ctx, []byte("a"), 3)
	require.False(t, estimate)
	require.Equal(t, []byte("2"), value)
	require.Equal(t, TxnVersion{Index: 2, Incarnation: 1}, version)

	// read closest version
	value, version, estimate = data.Read(ctx, []byte("b"), 3)
	require.False(t, estimate)
	require.Equal(t, []byte("2"), value)
	require.Equal(t, TxnVersion{Index: 2, Incarnation: 1}, version)

	// new incarnation overrides old
	data.Consolidate(ctx, TxnVersion{Index: 3, Incarnation: 2}, KV([]byte("a"), []byte("3-2")))
	value, version, estimate = data.Read(ctx, []byte("a"), 4)
	require.False(t, estimate)
	require.Equal(t, []byte("3-2"), value)
	require.Equal(t, TxnVersion{Index: 3, Incarnation: 2}, version)

	// read estimate
	data.ConvertWritesToEstimates(3)
	_, version, estimate = data.Read(ctx, []byte("a"), 4)
	require.True(t, estimate)
	require.Equal(t, TxnIndex(3), version.Index)

	// delete value
	data.deleteIndex(ctx, []byte("a"), 3)
	value, version, estimate = data.Read(ctx, []byte("a"), 4)
	require.False(t, estimate)
	require.Equal(t, []byte("2"), value)
	require.Equal(t, TxnVersion{Index: 2, Incarnation: 1}, version)

	data.deleteIndex(ctx, []byte("b"), 2)
	value, _, estimate = data.Read(ctx, []byte("b"), 4)
	require.False(t, estimate)
	require.Nil(t, value)
}

func TestReadErrConversion(t *testing.T) {
	err := fmt.Errorf("wrap: %w", ErrReadError{BlockingTxn: 1})
	var readErr ErrReadError
	require.True(t, errors.As(err, &readErr))
	require.Equal(t, TxnIndex(1), readErr.BlockingTxn)
}

type panickingStorage struct{}

func (panickingStorage) GetStoreType() storetypes.StoreType              { panic("not allowed") }
func (panickingStorage) Get([]byte) []byte                               { panic("not allowed") }
func (panickingStorage) Has([]byte) bool                                 { panic("validation re-probed storage") }
func (panickingStorage) Iterator(_, _ []byte) storetypes.Iterator        { panic("not allowed") }
func (panickingStorage) ReverseIterator(_, _ []byte) storetypes.Iterator { panic("not allowed") }

func TestHasReadValidation(t *testing.T) {
	ctx := context.Background()
	k := []byte("k")

	cases := []struct {
		name    string
		setup   func(*MVData, *MemDB)
		desc    HasDescriptor
		storage Storage // nil → use the populated MemDB
		valid   bool
	}{
		{
			name: "mvdata_version_flip_preserves_existence",
			setup: func(d *MVData, _ *MemDB) {
				d.Consolidate(ctx, TxnVersion{1, 0}, KV(k, []byte("v1")))
				d.Consolidate(ctx, TxnVersion{2, 0}, KV(k, []byte("v2")))
			},
			desc:  HasDescriptor{Key: k, Exists: true},
			valid: true,
		},
		{
			name: "mvdata_delete_flips_existence",
			setup: func(d *MVData, _ *MemDB) {
				d.Consolidate(ctx, TxnVersion{1, 0}, KV(k, []byte("v")))
				d.Consolidate(ctx, TxnVersion{2, 0}, KV(k, nil))
			},
			desc:  HasDescriptor{Key: k, Exists: true},
			valid: false,
		},
		{
			name:  "storage_fallback_matches",
			setup: func(_ *MVData, s *MemDB) { s.Set(k, []byte("pre")) },
			desc:  HasDescriptor{Key: k, Exists: true},
			valid: true,
		},
		{
			name:  "storage_fallback_mismatch",
			setup: func(_ *MVData, s *MemDB) { s.Set(k, []byte("pre")) },
			desc:  HasDescriptor{Key: k, Exists: false},
			valid: false,
		},
		{
			name:    "from_storage_skips_probe",
			desc:    HasDescriptor{Key: k, Exists: true, FromStorage: true},
			storage: panickingStorage{},
			valid:   true,
		},
		{
			name: "estimate_invalidates",
			setup: func(d *MVData, _ *MemDB) {
				d.Consolidate(ctx, TxnVersion{1, 0}, KV(k, []byte("v")))
				d.ConvertWritesToEstimates(1)
			},
			desc:  HasDescriptor{Key: k, Exists: true},
			valid: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := NewMVData(10)
			memdb := NewMemDB()
			if tc.setup != nil {
				tc.setup(data, memdb)
			}
			storage := Storage(memdb)
			if tc.storage != nil {
				storage = tc.storage
			}
			rs := &ReadSet{HasReads: []HasDescriptor{tc.desc}}
			require.Equal(t, tc.valid, data.ValidateReadSet(ctx, 3, rs, storage))
		})
	}
}

func TestSnapshot(t *testing.T) {
	ctx := context.Background()
	storage := NewMemDB()
	// initial value
	storage.Set([]byte("a"), []byte("0"))
	storage.Set([]byte("d"), []byte("0"))

	data := NewMVData(10)
	// read closest version
	data.Consolidate(ctx, TxnVersion{Index: 1, Incarnation: 1}, KV([]byte("a"), []byte("1")))
	data.Consolidate(ctx, TxnVersion{Index: 2, Incarnation: 1}, KV([]byte("a"), []byte("2")))
	data.Consolidate(ctx, TxnVersion{Index: 3, Incarnation: 1}, KV([]byte("a"), []byte("3")))
	data.Consolidate(ctx, TxnVersion{Index: 2, Incarnation: 1}, KV([]byte("a"), []byte("2"), []byte("b"), []byte("2"), []byte("d"), []byte("1")))
	// delete the key "d" in tx 3
	data.Consolidate(ctx, TxnVersion{Index: 3, Incarnation: 1}, KV([]byte("a"), []byte("3"), []byte("d"), nil))
	data.ConvertWritesToEstimates(2)

	require.Equal(t, []KVPair{
		{[]byte("a"), []byte("3")},
		{[]byte("b"), []byte("2")},
		{[]byte("d"), nil},
	}, data.Snapshot(ctx))

	data.SnapshotToStore(ctx, storage)
	require.Equal(t, []byte("3"), storage.Get([]byte("a")))
	require.Equal(t, []byte("2"), storage.Get([]byte("b")))
	require.Nil(t, storage.Get([]byte("d")))
	require.Equal(t, 2, storage.Len())
}
