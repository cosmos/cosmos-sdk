//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"bytes"
	"encoding/binary"

	"github.com/linxGnu/grocksdb"
)

// CreateTSComparator should behavior identical with RocksDB builtin timestamp comparator.
// We also use the same builtin comparator name so the builtin tools `ldb`/`sst_dump`
// can work with the database.
func CreateTSComparator() *grocksdb.Comparator {
	return grocksdb.NewComparatorWithTimestamp(
		"leveldb.BytewiseComparator.u64ts",
		TimestampSize,
		compare,
		compareTS,
		compareWithoutTS,
	)
}

// compareTS compares timestamp as little endian encoded integers.
//
// NOTICE: The behavior must be identical to RocksDB builtin comparator
// "leveldb.BytewiseComparator.u64ts".
func compareTS(bz1, bz2 []byte) int {
	ts1 := binary.LittleEndian.Uint64(bz1)
	ts2 := binary.LittleEndian.Uint64(bz2)

	switch {
	case ts1 < ts2:
		return -1

	case ts1 > ts2:
		return 1

	default:
		return 0
	}
}

// compare compares two internal keys with timestamp suffix, larger timestamp
// comes first.
//
// NOTICE: The behavior must be identical to RocksDB builtin comparator
// "leveldb.BytewiseComparator.u64ts".
func compare(a, b []byte) int {
	ret := compareWithoutTS(a, true, b, true)
	if ret != 0 {
		return ret
	}

	// Compare timestamp. For the same user key with different timestamps, larger
	// (newer) timestamp comes first, which means seek operation will try to find
	// a version less than or equal to the target version.
	return -compareTS(a[len(a)-TimestampSize:], b[len(b)-TimestampSize:])
}

// compareWithoutTS compares two internal keys without the timestamp part.
//
// NOTICE: the behavior must be identical to RocksDB builtin comparator
// "leveldb.BytewiseComparator.u64ts".
func compareWithoutTS(a []byte, aHasTS bool, b []byte, bHasTS bool) int {
	if aHasTS {
		a = a[:len(a)-TimestampSize]
	}
	if bHasTS {
		b = b[:len(b)-TimestampSize]
	}

	return bytes.Compare(a, b)
}
