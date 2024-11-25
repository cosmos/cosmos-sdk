package pebbledb

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cockroachdb/pebble"
)

// MVCCComparer returns a PebbleDB Comparer with encoding and decoding routines
// for MVCC control, used to compare and store versioned keys.
//
// Note: This Comparer implementation is largely based on PebbleDB's internal
// MVCC example, which can be found here:
// https://github.com/cockroachdb/pebble/blob/master/cmd/pebble/mvcc.go
var MVCCComparer = &pebble.Comparer{
	Name: "ss_pebbledb_comparator",

	Compare: MVCCKeyCompare,

	AbbreviatedKey: func(k []byte) uint64 {
		key, _, ok := SplitMVCCKey(k)
		if !ok {
			return 0
		}

		return pebble.DefaultComparer.AbbreviatedKey(key)
	},

	Equal: func(a, b []byte) bool {
		return MVCCKeyCompare(a, b) == 0
	},

	Separator: func(dst, a, b []byte) []byte {
		aKey, _, ok := SplitMVCCKey(a)
		if !ok {
			return append(dst, a...)
		}

		bKey, _, ok := SplitMVCCKey(b)
		if !ok {
			return append(dst, a...)
		}

		// if the keys are the same just return a
		if bytes.Equal(aKey, bKey) {
			return append(dst, a...)
		}

		n := len(dst)

		// MVCC key comparison uses bytes.Compare on the roachpb.Key, which is the
		// same semantics as pebble.DefaultComparer, so reuse the latter's Separator
		// implementation.
		dst = pebble.DefaultComparer.Separator(dst, aKey, bKey)

		// Did we pick a separator different than aKey? If we did not, we can't do
		// better than a.
		buf := dst[n:]
		if bytes.Equal(aKey, buf) {
			return append(dst[:n], a...)
		}

		// The separator is > aKey, so we only need to add the timestamp sentinel.
		return append(dst, 0)
	},

	ImmediateSuccessor: func(dst, a []byte) []byte {
		// The key `a` is guaranteed to be a bare prefix: It's a key without a version
		// — just a trailing 0-byte to signify the length of the version. For example
		// the user key "foo" is encoded as: "foo\0". We need to encode the immediate
		// successor to "foo", which in the natural byte ordering is "foo\0". Append
		// a single additional zero, to encode the user key "foo\0" with a zero-length
		// version.
		return append(append(dst, a...), 0)
	},

	Successor: func(dst, a []byte) []byte {
		aKey, _, ok := SplitMVCCKey(a)
		if !ok {
			return append(dst, a...)
		}

		n := len(dst)

		// MVCC key comparison uses bytes.Compare on the roachpb.Key, which is the
		// same semantics as pebble.DefaultComparer, so reuse the latter's Successor
		// implementation.
		dst = pebble.DefaultComparer.Successor(dst, aKey)

		// Did we pick a successor different than aKey? If we did not, we can't do
		// better than a.
		buf := dst[n:]
		if bytes.Equal(aKey, buf) {
			return append(dst[:n], a...)
		}

		// The successor is > aKey, so we only need to add the timestamp sentinel.
		return append(dst, 0)
	},

	FormatKey: func(k []byte) fmt.Formatter {
		return mvccKeyFormatter{key: k}
	},

	Split: func(k []byte) int {
		key, _, ok := SplitMVCCKey(k)
		if !ok {
			return len(k)
		}

		// This matches the behavior of libroach/KeyPrefix. RocksDB requires that
		// keys generated via a SliceTransform be comparable with normal encoded
		// MVCC keys. Encoded MVCC keys have a suffix indicating the number of
		// bytes of timestamp data. MVCC keys without a timestamp have a suffix of
		// 0. We're careful in EncodeKey to make sure that the user-key always has
		// a trailing 0. If there is no timestamp this falls out naturally. If
		// there is a timestamp we prepend a 0 to the encoded timestamp data.
		return len(key) + 1
	},
}

type mvccKeyFormatter struct {
	key []byte
}

func (f mvccKeyFormatter) Format(s fmt.State, verb rune) {
	k, vBz, ok := SplitMVCCKey(f.key)
	if ok {
		v, _ := decodeUint64Ascending(vBz)
		fmt.Fprintf(s, "%s/%d", k, v)
	} else {
		fmt.Fprintf(s, "%s", f.key)
	}
}

// SplitMVCCKey accepts an MVCC key and returns the "user" key, the MVCC version,
// and a boolean indicating if the provided key is an MVCC key.
//
// Note, internally, we must make a copy of the provided mvccKey argument, which
// typically comes from the Key() method as it's not safe.
func SplitMVCCKey(mvccKey []byte) (key, version []byte, ok bool) {
	if len(mvccKey) == 0 {
		return nil, nil, false
	}

	mvccKeyCopy := bytes.Clone(mvccKey)

	n := len(mvccKeyCopy) - 1
	tsLen := int(mvccKeyCopy[n])
	if n < tsLen {
		return nil, nil, false
	}

	key = mvccKeyCopy[:n-tsLen]
	if tsLen > 0 {
		version = mvccKeyCopy[n-tsLen+1 : n]
	}

	return key, version, true
}

// MVCCKeyCompare compares two MVCC keys.
func MVCCKeyCompare(a, b []byte) int {
	aEnd := len(a) - 1
	bEnd := len(b) - 1
	if aEnd < 0 || bEnd < 0 {
		// This should never happen unless there is some sort of corruption of
		// the keys. This is a little bizarre, but the behavior exactly matches
		// engine/db.cc:DBComparator.
		return bytes.Compare(a, b)
	}

	// Compute the index of the separator between the key and the timestamp.
	aSep := aEnd - int(a[aEnd])
	bSep := bEnd - int(b[bEnd])
	if aSep < 0 || bSep < 0 {
		// This should never happen unless there is some sort of corruption of
		// the keys. This is a little bizarre, but the behavior exactly matches
		// engine/db.cc:DBComparator.
		return bytes.Compare(a, b)
	}

	// compare the "user key" part of the key
	if c := bytes.Compare(a[:aSep], b[:bSep]); c != 0 {
		return c
	}

	// compare the timestamp part of the key
	aTS := a[aSep:aEnd]
	bTS := b[bSep:bEnd]
	if len(aTS) == 0 {
		if len(bTS) == 0 {
			return 0
		}
		return -1
	} else if len(bTS) == 0 {
		return 1
	}

	return bytes.Compare(aTS, bTS)
}

// MVCCEncode encodes a key and version into an MVCC format.
// The format is: <key>\x00[<version>]<#version-bytes>
// If the version is 0, only the key and a null byte are encoded.
func MVCCEncode(key []byte, version uint64) (dst []byte) {
	dst = append(dst, key...)
	dst = append(dst, 0)

	if version != 0 {
		extra := byte(1 + 8)
		dst = encodeUint64Ascending(dst, version)
		dst = append(dst, extra)
	}

	return dst
}

// encodeUint64Ascending encodes the uint64 value using a big-endian 8 byte
// representation. The bytes are appended to the supplied buffer and
// the final buffer is returned.
func encodeUint64Ascending(dst []byte, v uint64) []byte {
	return append(
		dst,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v),
	)
}

// decodeUint64Ascending decodes a uint64 from the input buffer, treating
// the input as a big-endian 8 byte uint64 representation. The decoded uint64 is
// returned.
func decodeUint64Ascending(b []byte) (uint64, error) {
	if len(b) < 8 {
		return 0, fmt.Errorf("insufficient bytes to decode uint64 int value; expected 8; got %d", len(b))
	}

	v := binary.BigEndian.Uint64(b)
	return v, nil
}
