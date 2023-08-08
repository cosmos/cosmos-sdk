package pebbledb

import (
	"bytes"
	"encoding/binary"

	"github.com/cockroachdb/pebble"
)

// MVCCComparer returns a PebbleDB Comparer with encoding and decoding routines
// for MVCC control, used to compare and store versioned data.
var MVCCComparer = &pebble.Comparer{
	Name:    "ss_pebbledb_comparator",
	Compare: MVCCKeyCompare,
	// Equal: func(a, b []byte) bool {
	// 	return mvccCompare(a, b) == 0
	// },
	// AbbreviatedKey: func(k []byte) uint64 {
	// 	key, _, ok := mvccSplitKey(k)
	// 	if !ok {
	// 		return 0
	// 	}
	// 	return pebble.DefaultComparer.AbbreviatedKey(key)
	// },

	// Separator: func(dst, a, b []byte) []byte {
	// 	aKey, _, ok := mvccSplitKey(a)
	// 	if !ok {
	// 		return append(dst, a...)
	// 	}
	// 	bKey, _, ok := mvccSplitKey(b)
	// 	if !ok {
	// 		return append(dst, a...)
	// 	}
	// 	// If the keys are the same just return a.
	// 	if bytes.Equal(aKey, bKey) {
	// 		return append(dst, a...)
	// 	}
	// 	n := len(dst)
	// 	// MVCC key comparison uses bytes.Compare on the roachpb.Key, which is the same semantics as
	// 	// pebble.DefaultComparer, so reuse the latter's Separator implementation.
	// 	dst = pebble.DefaultComparer.Separator(dst, aKey, bKey)
	// 	// Did it pick a separator different than aKey -- if it did not we can't do better than a.
	// 	buf := dst[n:]
	// 	if bytes.Equal(aKey, buf) {
	// 		return append(dst[:n], a...)
	// 	}
	// 	// The separator is > aKey, so we only need to add the timestamp sentinel.
	// 	return append(dst, 0)
	// },

	// Successor: func(dst, a []byte) []byte {
	// 	aKey, _, ok := mvccSplitKey(a)
	// 	if !ok {
	// 		return append(dst, a...)
	// 	}
	// 	n := len(dst)
	// 	// MVCC key comparison uses bytes.Compare on the roachpb.Key, which is the same semantics as
	// 	// pebble.DefaultComparer, so reuse the latter's Successor implementation.
	// 	dst = pebble.DefaultComparer.Successor(dst, aKey)
	// 	// Did it pick a successor different than aKey -- if it did not we can't do better than a.
	// 	buf := dst[n:]
	// 	if bytes.Equal(aKey, buf) {
	// 		return append(dst[:n], a...)
	// 	}
	// 	// The successor is > aKey, so we only need to add the timestamp sentinel.
	// 	return append(dst, 0)
	// },

	// Split: func(k []byte) int {
	// 	key, _, ok := mvccSplitKey(k)
	// 	if !ok {
	// 		return len(k)
	// 	}
	// 	// This matches the behavior of libroach/KeyPrefix. RocksDB requires that
	// 	// keys generated via a SliceTransform be comparable with normal encoded
	// 	// MVCC keys. Encoded MVCC keys have a suffix indicating the number of
	// 	// bytes of timestamp data. MVCC keys without a timestamp have a suffix of
	// 	// 0. We're careful in EncodeKey to make sure that the user-key always has
	// 	// a trailing 0. If there is no timestamp this falls out naturally. If
	// 	// there is a timestamp we prepend a 0 to the encoded timestamp data.
	// 	return len(key) + 1
	// },
}

// MVCCKeyCompare compares PebbleDB versioned store keys, which are MVCC timestamps.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b.
func MVCCKeyCompare(a, b []byte) int {
	keyA, tsA, okA := SplitMVCCKey(a)
	keyB, tsB, okB := SplitMVCCKey(b)

	if !okA || !okB {
		return bytes.Compare(a, b)
	}

	// compare the "user key" part of the key
	if res := bytes.Compare(keyA, keyB); res != 0 {
		return res
	}

	if len(tsA) == 0 {
		if len(tsB) == 0 {
			return 0
		}

		return -1
	} else if len(tsB) == 0 {
		return 1
	}

	return bytes.Compare(tsA, tsB)
}

func SplitMVCCKey(mvccKey []byte) (key []byte, ts []byte, ok bool) {
	if len(mvccKey) == 0 {
		return nil, nil, false
	}

	tsLen := int(mvccKey[len(mvccKey)-1])
	keyPartEnd := len(mvccKey) - 1 - tsLen
	if keyPartEnd < 0 {
		return nil, nil, false
	}

	key = mvccKey[:keyPartEnd]
	if tsLen > 0 {
		ts = mvccKey[keyPartEnd+1 : len(mvccKey)-1]
	}

	return key, ts, true
}

// MVCCEncode encodes a key for MVCC control. Note, adapted from PebbleDB.
//
// <key>\x00[<version>]<#version-bytes>
func MVCCEncode(key []byte, version uint64) (dst []byte) {
	dst = append(dst, key...)
	dst = append(dst, 0)

	if version != 0 {
		extra := byte(1 + 8)

		var versionBz [VersionSize]byte
		binary.LittleEndian.PutUint64(versionBz[:], version)

		dst = append(dst, versionBz[:]...)
		dst = append(dst, extra)
	}

	return dst
}
