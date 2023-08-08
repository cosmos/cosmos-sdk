package pebbledb

import "github.com/cockroachdb/pebble"

// MVCCComparer returns a PebbleDB Comparer with encoding and decoding routines
// for MVCC control, used to compare and store versioned data.
var MVCCComparer = &pebble.Comparer{}

// mvccEncode encodes a key for MVCC control. Note, adapted from PebbleDB.
//
// <key>\x00[<version>]<#version-bytes>
func mvccEncode(key []byte, version uint64) (dst []byte) {
	dst = append(dst, key...)
	dst = append(dst, 0)

	if version != 0 {
		extra := byte(1 + 8)
		dst = encodeUint64Ascending(dst, version)
		dst = append(dst, extra)
	}

	return dst
}

func encodeUint64Ascending(b []byte, v uint64) []byte {
	return append(b,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
