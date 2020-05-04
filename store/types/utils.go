package types

import (
	"bytes"

	tmkv "github.com/tendermint/tendermint/libs/kv"
)

// Iterator over all the keys with a certain prefix in ascending order
func KVStorePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return kvs.Iterator(prefix, PrefixEndBytes(prefix))
}

// Iterator over all the keys with a certain prefix in descending order.
func KVStoreReversePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return kvs.ReverseIterator(prefix, PrefixEndBytes(prefix))
}

// DiffKVStores compares two KVstores and returns all the key/value pairs
// that differ from one another. It also skips value comparison for a set of provided prefixes.
func DiffKVStores(a KVStore, b KVStore, prefixesToSkip [][]byte) (kvAs, kvBs []tmkv.Pair) {
	iterA := a.Iterator(nil, nil)

	defer iterA.Close()

	iterB := b.Iterator(nil, nil)

	defer iterB.Close()

	for {
		if !iterA.Valid() && !iterB.Valid() {
			return kvAs, kvBs
		}

		var kvA, kvB tmkv.Pair
		if iterA.Valid() {
			kvA = tmkv.Pair{Key: iterA.Key(), Value: iterA.Value()}

			iterA.Next()
		}

		if iterB.Valid() {
			kvB = tmkv.Pair{Key: iterB.Key(), Value: iterB.Value()}

			iterB.Next()
		}

		compareValue := true

		for _, prefix := range prefixesToSkip {
			// Skip value comparison if we matched a prefix
			if bytes.HasPrefix(kvA.Key, prefix) || bytes.HasPrefix(kvB.Key, prefix) {
				compareValue = false
				break
			}
		}

		if compareValue && (!bytes.Equal(kvA.Key, kvB.Key) || !bytes.Equal(kvA.Value, kvB.Value)) {
			kvAs = append(kvAs, kvA)
			kvBs = append(kvBs, kvB)
		}
	}
}

// PrefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func PrefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		}

		end = end[:len(end)-1]

		if len(end) == 0 {
			end = nil
			break
		}
	}

	return end
}

// InclusiveEndBytes returns the []byte that would end a
// range query such that the input would be included
func InclusiveEndBytes(inclusiveBytes []byte) []byte {
	return append(inclusiveBytes, byte(0x00))
}
