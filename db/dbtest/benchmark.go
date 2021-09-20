package dbtest

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-sdk/db"
)

func Int64ToBytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func BenchmarkRangeScans(b *testing.B, db dbm.DBReadWriter, dbSize int64) {
	b.StopTimer()

	rangeSize := int64(10000)
	if dbSize < rangeSize {
		b.Errorf("db size %v cannot be less than range size %v", dbSize, rangeSize)
	}

	for i := int64(0); i < dbSize; i++ {
		bytes := Int64ToBytes(i)
		err := db.Set(bytes, bytes)
		if err != nil {
			// require.NoError() is very expensive (according to profiler), so check manually
			b.Fatal(b, err)
		}
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		start := rand.Int63n(dbSize - rangeSize) // nolint: gosec
		end := start + rangeSize
		iter, err := db.Iterator(Int64ToBytes(start), Int64ToBytes(end))
		require.NoError(b, err)
		count := 0
		for iter.Next() {
			count++
		}
		iter.Close()
		require.EqualValues(b, rangeSize, count)
	}
}

func BenchmarkRandomReadsWrites(b *testing.B, db dbm.DBReadWriter) {
	b.StopTimer()

	// create dummy data
	const numItems = int64(1000000)
	internal := map[int64]int64{}
	for i := 0; i < int(numItems); i++ {
		internal[int64(i)] = int64(0)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		{
			idx := rand.Int63n(numItems) // nolint: gosec
			internal[idx]++
			val := internal[idx]
			idxBytes := Int64ToBytes(idx)
			valBytes := Int64ToBytes(val)
			err := db.Set(idxBytes, valBytes)
			if err != nil {
				// require.NoError() is very expensive (according to profiler), so check manually
				b.Fatal(b, err)
			}
		}

		{
			idx := rand.Int63n(numItems) // nolint: gosec
			valExp := internal[idx]
			idxBytes := Int64ToBytes(idx)
			valBytes, err := db.Get(idxBytes)
			if err != nil {
				b.Fatal(b, err)
			}
			if valExp == 0 {
				if !bytes.Equal(valBytes, nil) {
					b.Errorf("Expected %v for %v, got %X", nil, idx, valBytes)
					break
				}
			} else {
				if len(valBytes) != 8 {
					b.Errorf("Expected length 8 for %v, got %X", idx, valBytes)
					break
				}
				valGot := BytesToInt64(valBytes)
				if valExp != valGot {
					b.Errorf("Expected %v for %v, got %v", valExp, idx, valGot)
					break
				}
			}
		}

	}
}
