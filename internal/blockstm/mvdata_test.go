package blockstm

import (
	"errors"
	"fmt"
	"testing"

	"github.com/test-go/testify/require"
)

func TestEmptyMVData(t *testing.T) {
	data := NewMVData()
	value, _, estimate := data.Read([]byte("a"), 1)
	require.False(t, estimate)
	require.Nil(t, value)
}

func TestMVData(t *testing.T) {
	data := NewMVData()

	// read closest version
	data.Write([]byte("a"), []byte("1"), TxnVersion{Index: 1, Incarnation: 1})
	data.Write([]byte("a"), []byte("2"), TxnVersion{Index: 2, Incarnation: 1})
	data.Write([]byte("a"), []byte("3"), TxnVersion{Index: 3, Incarnation: 1})
	data.Write([]byte("b"), []byte("2"), TxnVersion{Index: 2, Incarnation: 1})

	// read closest version
	value, _, estimate := data.Read([]byte("a"), 1)
	require.False(t, estimate)
	require.Nil(t, value)

	// read closest version
	value, version, estimate := data.Read([]byte("a"), 4)
	require.False(t, estimate)
	require.Equal(t, []byte("3"), value)
	require.Equal(t, TxnVersion{Index: 3, Incarnation: 1}, version)

	// read closest version
	value, version, estimate = data.Read([]byte("a"), 3)
	require.False(t, estimate)
	require.Equal(t, []byte("2"), value)
	require.Equal(t, TxnVersion{Index: 2, Incarnation: 1}, version)

	// read closest version
	value, version, estimate = data.Read([]byte("b"), 3)
	require.False(t, estimate)
	require.Equal(t, []byte("2"), value)
	require.Equal(t, TxnVersion{Index: 2, Incarnation: 1}, version)

	// new incarnation overrides old
	data.Write([]byte("a"), []byte("3-2"), TxnVersion{Index: 3, Incarnation: 2})
	value, version, estimate = data.Read([]byte("a"), 4)
	require.False(t, estimate)
	require.Equal(t, []byte("3-2"), value)
	require.Equal(t, TxnVersion{Index: 3, Incarnation: 2}, version)

	// read estimate
	data.WriteEstimate([]byte("a"), 3)
	_, version, estimate = data.Read([]byte("a"), 4)
	require.True(t, estimate)
	require.Equal(t, TxnIndex(3), version.Index)

	// delete value
	data.Delete([]byte("a"), 3)
	value, version, estimate = data.Read([]byte("a"), 4)
	require.False(t, estimate)
	require.Equal(t, []byte("2"), value)
	require.Equal(t, TxnVersion{Index: 2, Incarnation: 1}, version)

	data.Delete([]byte("b"), 2)
	value, _, estimate = data.Read([]byte("b"), 4)
	require.False(t, estimate)
	require.Nil(t, value)
}

func TestReadErrConversion(t *testing.T) {
	err := fmt.Errorf("wrap: %w", ErrReadError{BlockingTxn: 1})
	var readErr ErrReadError
	require.True(t, errors.As(err, &readErr))
	require.Equal(t, TxnIndex(1), readErr.BlockingTxn)
}

func TestSnapshot(t *testing.T) {
	storage := NewMemDB()
	// initial value
	storage.Set([]byte("a"), []byte("0"))
	storage.Set([]byte("d"), []byte("0"))

	data := NewMVData()
	// read closest version
	data.Write([]byte("a"), []byte("1"), TxnVersion{Index: 1, Incarnation: 1})
	data.Write([]byte("a"), []byte("2"), TxnVersion{Index: 2, Incarnation: 1})
	data.Write([]byte("a"), []byte("3"), TxnVersion{Index: 3, Incarnation: 1})
	data.Write([]byte("b"), []byte("2"), TxnVersion{Index: 2, Incarnation: 1})
	data.Write([]byte("d"), []byte("1"), TxnVersion{Index: 2, Incarnation: 1})
	// delete the key "d" in tx 3
	data.Write([]byte("d"), nil, TxnVersion{Index: 3, Incarnation: 1})
	data.WriteEstimate([]byte("c"), 2)

	require.Equal(t, []KVPair{
		{[]byte("a"), []byte("3")},
		{[]byte("b"), []byte("2")},
		{[]byte("d"), nil},
	}, data.Snapshot())

	data.SnapshotToStore(storage)
	require.Equal(t, []byte("3"), storage.Get([]byte("a")))
	require.Equal(t, []byte("2"), storage.Get([]byte("b")))
	require.Nil(t, storage.Get([]byte("d")))
	require.Equal(t, 2, storage.Len())
}

func TestValueBasedValidationAcceptsSameValueDifferentVersion(t *testing.T) {
	data := NewMVData()

	// Txn 3 reads "v" from Txn 1.
	rs := &ReadSet{
		Reads: []ReadDescriptor{{Key: []byte("k"), Version: TxnVersion{Index: 1, Incarnation: 1}, Captured: []byte("v")}},
	}

	// Both Txn 1 and Txn 2 write "v". Closest dependency for Txn 3 becomes Txn 2.
	data.Write([]byte("k"), []byte("v"), TxnVersion{Index: 1, Incarnation: 1})
	data.Write([]byte("k"), []byte("v"), TxnVersion{Index: 2, Incarnation: 1})

	// Validates: Version mismatch (exp 1, got 2) ignored matching value.
	require.True(t, data.ValidateReadSet(3, rs))

	// Invalidates: Value mismatch.
	rs.Reads[0].Captured = []byte("other")
	require.False(t, data.ValidateReadSet(3, rs))
}

func TestHasValidation(t *testing.T) {
	data := NewMVData()

	// Txn 1 writes "k".
	data.Write([]byte("k"), []byte("v"), TxnVersion{Index: 1, Incarnation: 0})

	// Txn 3 checks Has("k") == true.
	rs := &ReadSet{
		Reads: []ReadDescriptor{{Key: []byte("k"), Has: true, ExistsExpected: true}},
	}

	// Validates: "k" exists (from Txn 1).
	require.True(t, data.ValidateReadSet(3, rs))

	// Txn 2 also writes "k". Txn 3 now depends on Txn 2.
	data.Write([]byte("k"), []byte("v2"), TxnVersion{Index: 2, Incarnation: 0})

	// Validates: "k" still exists (from Txn 2), even if version changed.
	require.True(t, data.ValidateReadSet(3, rs))

	// Txn 2 deletes "k".
	data.Write([]byte("k"), nil, TxnVersion{Index: 2, Incarnation: 1})

	// Invalidates: "k" does not exist anymore.
	require.False(t, data.ValidateReadSet(3, rs))
}
