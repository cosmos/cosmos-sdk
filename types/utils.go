package types

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/types/kv"
)

var (
	// This is set at compile time. Could be cleveldb, defaults is goleveldb.
	DBBackend = "" // Deprecated: Use tendermint config's DBBackend value instead.
	backend   = dbm.GoLevelDBBackend
)

func init() {
	if len(DBBackend) != 0 {
		backend = dbm.BackendType(DBBackend)
	}
}

// SortedJSON takes any JSON and returns it sorted by keys. Also, all white-spaces
// are removed.
// This method can be used to canonicalize JSON to be returned by GetSignBytes,
// e.g. for the ledger integration.
// If the passed JSON isn't valid it will return an error.
func SortJSON(toSortJSON []byte) ([]byte, error) {
	var c interface{}
	err := json.Unmarshal(toSortJSON, &c)
	if err != nil {
		return nil, err
	}
	js, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return js, nil
}

// MustSortJSON is like SortJSON but panic if an error occurs, e.g., if
// the passed JSON isn't valid.
func MustSortJSON(toSortJSON []byte) []byte {
	js, err := SortJSON(toSortJSON)
	if err != nil {
		panic(err)
	}
	return js
}

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
	return []byte(t.UTC().Round(0).Format(SortableTimeFormat))
}

// Parses a []byte encoded using FormatTimeKey back into a time.Time
func ParseTimeBytes(bz []byte) (time.Time, error) {
	str := string(bz)
	t, err := time.Parse(SortableTimeFormat, str)
	if err != nil {
		return t, err
	}
	return t.UTC().Round(0), nil
}

// NewLevelDB instantiate a new LevelDB instance according to DBBackend.
//
// Deprecated: Use NewDB (from "github.com/tendermint/tm-db") instead. Suggested backendType is tendermint config's DBBackend value.
func NewLevelDB(name, dir string) (db dbm.DB, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("couldn't create db: %v", r)
		}
	}()

	return dbm.NewDB(name, backend, dir)
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

// AppendLengthPrefixedBytes combines the slices of bytes to one slice of bytes.
func AppendLengthPrefixedBytes(args ...[]byte) []byte {
	length := 0
	for _, v := range args {
		length += len(v)
	}
	res := make([]byte, length)

	length = 0
	for _, v := range args {
		copy(res[length:length+len(v)], v)
		length += len(v)
	}

	return res
}

// ParseLengthPrefixedBytes panics when store key length is not equal to the given length.
func ParseLengthPrefixedBytes(key []byte, startIndex int, sliceLength int) ([]byte, int) {
	neededLength := startIndex + sliceLength
	endIndex := neededLength - 1
	kv.AssertKeyAtLeastLength(key, neededLength)
	byteSlice := key[startIndex:neededLength]

	return byteSlice, endIndex
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
