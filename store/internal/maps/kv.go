package maps

import (
	"bytes"
	"sort"
)

type (
	KVPair struct {
		Key   []byte
		Value []byte
	}

	KVPairs struct {
		Pairs []KVPair
	}
)

// NewKVPair takes in a key and value and creates a kv.Pair
// wrapped in the local extension KVPair
func NewKVPair(key, value []byte) KVPair {
	return KVPair(KVPair{
		Key:   key,
		Value: value,
	})
}

func (kvs KVPairs) Len() int { return len(kvs.Pairs) }
func (kvs KVPairs) Less(i, j int) bool {
	switch bytes.Compare(kvs.Pairs[i].Key, kvs.Pairs[j].Key) {
	case -1:
		return true

	case 0:
		return bytes.Compare(kvs.Pairs[i].Value, kvs.Pairs[j].Value) < 0

	case 1:
		return false

	default:
		panic("invalid comparison result")
	}
}

func (kvs KVPairs) Swap(i, j int) { kvs.Pairs[i], kvs.Pairs[j] = kvs.Pairs[j], kvs.Pairs[i] }

// Sort invokes sort.Sort on kvs.
func (kvs KVPairs) Sort() { sort.Sort(kvs) }
