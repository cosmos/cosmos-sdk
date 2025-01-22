package simapp

import (
	"bytes"
	"fmt"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/store"
	storev2 "cosmossdk.io/store/v2"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

func AssertEqualStores(tb testing.TB, srcRootStore, otherRootStore storev2.RootStore, storeKeys []string, storeDecoders simulation.StoreDecoderRegistry, skipPrefixes map[string][][]byte) {
	tb.Helper()
	srcVersion, srcStores, err := srcRootStore.StateLatest()
	require.NoError(tb, err)
	otherVersion, otherStores, err := otherRootStore.StateLatest()
	require.NoError(tb, err)
	require.Equal(tb, srcVersion, otherVersion, "store versions do not match")
	for _, storeKey := range storeKeys {
		srcReader, err := srcStores.GetReader([]byte(storeKey))
		require.NoError(tb, err)
		otherReader, err := otherStores.GetReader([]byte(storeKey))
		require.NoError(tb, err)
		failedKVAs, failedKVBs := DiffKVStores(tb, storeKey, srcReader, otherReader, skipPrefixes[storeKey])
		if !assert.Empty(tb, len(failedKVAs), storeKey+": "+GetSimulationLog(storeKey, storeDecoders, failedKVAs, failedKVBs)) {
			for i, v := range failedKVAs {
				tb.Logf("store mismatch: %q\n %q\n", v, failedKVBs[i])
			}
			tb.FailNow()
		}
		tb.Logf("compared %d different key/value pairs for %s\n", len(failedKVAs), storeKey)
	}
}

func DiffKVStores(tb testing.TB, storeKey string, a, b store.Reader, prefixesToSkip [][]byte) (diffA, diffB []kv.Pair) {
	tb.Helper()
	iterA, err := a.Iterator(nil, nil)
	require.NoError(tb, err, storeKey)
	defer iterA.Close()

	iterB, err := b.Iterator(nil, nil)
	require.NoError(tb, err, storeKey)
	defer iterB.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	kvAs := make([]kv.Pair, 0)
	go func() {
		defer wg.Done()
		kvAs = getKVPairs(iterA, prefixesToSkip)
	}()

	wg.Add(1)
	kvBs := make([]kv.Pair, 0)
	go func() {
		defer wg.Done()
		kvBs = getKVPairs(iterB, prefixesToSkip)
	}()

	wg.Wait()

	if len(kvAs) != len(kvBs) {
		fmt.Printf("%q:: KV stores are different: %d key/value pairs in store A and %d key/value pairs in store B\n", storeKey, len(kvAs), len(kvBs))
	}

	return getDiffFromKVPair(kvAs, kvBs)
}

func getKVPairs(iter store.Iterator, prefixesToSkip [][]byte) (kvs []kv.Pair) {
	for iter.Valid() {
		key, value := iter.Key(), iter.Value()

		// do not add the KV pair if the key is prefixed to be skipped.
		skip := slices.ContainsFunc(prefixesToSkip, func(prefix []byte) bool {
			return bytes.HasPrefix(key, prefix)
		})
		if !skip {
			kvs = append(kvs, kv.Pair{Key: key, Value: value})
		}
		iter.Next()
	}

	return kvs
}

// getDiffFromKVPair compares two KVstores and returns all the key/value pairs
func getDiffFromKVPair(kvAs, kvBs []kv.Pair) (diffA, diffB []kv.Pair) {
	// we assume that kvBs is equal or larger than kvAs
	// if not, we swap the two
	if len(kvAs) > len(kvBs) {
		kvAs, kvBs = kvBs, kvAs
		// we need to swap the diffA and diffB as well
		defer func() {
			diffA, diffB = diffB, diffA
		}()
	}

	// in case kvAs is empty we can return early
	// since there is nothing to compare
	// if kvAs == kvBs, then diffA and diffB will be empty
	if len(kvAs) == 0 {
		return []kv.Pair{}, kvBs
	}

	index := make(map[string][]byte, len(kvBs))
	for _, kv := range kvBs {
		index[string(kv.Key)] = kv.Value
	}

	for _, kvA := range kvAs {
		if kvBValue, ok := index[string(kvA.Key)]; !ok {
			diffA = append(diffA, kvA)
			diffB = append(diffB, kv.Pair{Key: kvA.Key}) // the key is missing from kvB so we append a pair with an empty value
		} else if !bytes.Equal(kvA.Value, kvBValue) {
			diffA = append(diffA, kvA)
			diffB = append(diffB, kv.Pair{Key: kvA.Key, Value: kvBValue})
		} else {
			// values are equal, so we remove the key from the index
			delete(index, string(kvA.Key))
		}
	}

	// add the remaining keys from kvBs
	for key, value := range index {
		diffA = append(diffA, kv.Pair{Key: []byte(key)}) // the key is missing from kvA so we append a pair with an empty value
		diffB = append(diffB, kv.Pair{Key: []byte(key), Value: value})
	}

	return diffA, diffB
}

func defaultSimLogDecoder(kvA, kvB kv.Pair) string {
	return fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvA.Key, kvA.Value, kvB.Key, kvB.Value)
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr simulation.StoreDecoderRegistry, kvAs, kvBs []kv.Pair) (log string) {
	decoder := defaultSimLogDecoder
	if dec, ok := sdr[storeName]; ok {
		decoder = dec
	}
	for i := 0; i < len(kvAs); i++ {
		if len(kvAs[i].Value) == 0 && len(kvBs[i].Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}
		log += decoder(kvAs[i], kvBs[i])
	}

	return log
}
