package kv

import (
	"bytes"
	"sort"
)

//----------------------------------------
// KVPair

/*
Defined in types.proto

type Pair struct {
	Key   []byte
	Value []byte
}
*/

type Pairs []Pair

// Sorting
func (kvs Pairs) Len() int { return len(kvs) }
func (kvs Pairs) Less(i, j int) bool {
	switch bytes.Compare(kvs[i].Key, kvs[j].Key) {
	case -1:
		return true
	case 0:
		return bytes.Compare(kvs[i].Value, kvs[j].Value) < 0
	case 1:
		return false
	default:
		panic("invalid comparison result")
	}
}
func (kvs Pairs) Swap(i, j int) { kvs[i], kvs[j] = kvs[j], kvs[i] }
func (kvs Pairs) Sort()         { sort.Sort(kvs) }
