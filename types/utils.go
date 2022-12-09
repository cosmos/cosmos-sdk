package types

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types/kv"
)

// Uint64ToBigEndian - marshals uint64 to a bigendian byte slice so it can be sorted
func Uint64ToBigEndian(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

// BigEndianToUint64 returns an uint64 from big endian encoded bytes. If encoding
// is empty, zero is returned.
func BigEndianToUint64(bz []byte) uint64 {
	if len(bz) == 0 {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

// Slight modification of the RFC3339Nano but it right pads all zeros and drops the time zone info
const SortableTimeFormat = "2006-01-02T15:04:05.000000000"

// Formats a time.Time into a []byte that can be sorted
func FormatTimeBytes(t time.Time) []byte {
	return []byte(FormatTimeString(t))
}

// Formats a time.Time into a string
func FormatTimeString(t time.Time) string {
	return t.UTC().Round(0).Format(SortableTimeFormat)
}

// Parses a []byte encoded using FormatTimeKey back into a time.Time
func ParseTimeBytes(bz []byte) (time.Time, error) {
	return ParseTime(bz)
}

// Parses an encoded type using FormatTimeKey back into a time.Time
func ParseTime(t any) (time.Time, error) {
	var (
		result time.Time
		err    error
	)

	switch t := t.(type) {
	case time.Time:
		result, err = t, nil
	case []byte:
		result, err = time.Parse(SortableTimeFormat, string(t))
	case string:
		result, err = time.Parse(SortableTimeFormat, t)
	default:
		return time.Time{}, fmt.Errorf("unexpected type %T", t)
	}

	if err != nil {
		return result, err
	}

	return result.UTC().Round(0), nil
}

// copy bytes
func CopyBytes(bz []byte) (ret []byte) {
	if bz == nil {
		return nil
	}
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}

// SliceContains implements a generic function for checking if a slice contains
// a certain value.
func SliceContains[T comparable](elements []T, v T) bool {
	for _, s := range elements {
		if v == s {
			return true
		}
	}

	return false
}
