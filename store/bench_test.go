package store

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/db/badgerdb"
	"github.com/cosmos/cosmos-sdk/db/rocksdb"
	storev1 "github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/types"
	storev2 "github.com/cosmos/cosmos-sdk/store/v2/flat"
	dbm "github.com/tendermint/tm-db"
)

var (
	cacheSize = 100
)

func randBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, _ = crand.Read(b)
	return b
}

type percentages struct {
	has    uint
	get    uint
	set    uint
	delete uint
}

type counts struct {
	has    int
	get    int
	set    int
	delete int
}

func generateSampledPercentages() []*percentages {
	var sampledPercentages []*percentages
	sampleX := &percentages{has: 5, get: 65, set: 30, delete: 0}
	sampledPercentages = append(sampledPercentages, sampleX)
	for a := uint(0); a < 100; a += 20 {
		for b := uint(0); b < 100; b += 20 {
			if a+b > 100 {
				break
			} else {
				for c := uint(0); c < 100; c += 20 {
					if a+b+c > 100 {
						break
					} else {
						sample := percentages{
							has:    a,
							get:    b,
							set:    c,
							delete: 100 - a - b - c,
						}
						sampledPercentages = append(sampledPercentages, &sample)
					}
				}
			}
		}
	}
	return sampledPercentages
}

type benchmark struct {
	name        string
	percentages *percentages
	dbType      dbm.BackendType
	counts      *counts
}

func generateBenchmarks(dbBackendTypes []dbm.BackendType, sampledPercentages []*percentages, sampledCounts []*counts) []benchmark {
	var benchmarks []benchmark
	for _, dbType := range dbBackendTypes {
		if len(sampledPercentages) > 0 {
			for _, p := range sampledPercentages {
				name := fmt.Sprintf("r-%s-%d-%d-%d-%d", dbType, p.has, p.get, p.set, p.delete)
				benchmarks = append(benchmarks, benchmark{name: name, percentages: p, dbType: dbType, counts: (*counts)(nil)})
			}
		} else if len(sampledCounts) > 0 {
			for _, c := range sampledCounts {
				name := fmt.Sprintf("d-%s-%d-%d-%d-%d", dbType, c.has, c.get, c.set, c.delete)
				benchmarks = append(benchmarks, benchmark{name: name, percentages: (*percentages)(nil), dbType: dbType, counts: c})
			}
		}
	}
	return benchmarks
}

type store interface {
	Has(key []byte) bool
	Get(key []byte) []byte
	Set(key []byte, value []byte)
	Delete(key []byte)
	Commit() types.CommitID
}

func sampleOperation(p *percentages) (string, error) {
	ops := []string{"Has", "Get", "Set", "Delete"}
	thresholds := []uint{p.has, p.has + p.get, p.has + p.get + p.set}
	r := rand.Intn(100)
	if r >= int(thresholds[2]) {
		return ops[3], nil
	} else {
		for i := 0; i < len(thresholds); i++ {
			if r < int(thresholds[i]) {
				return ops[i], nil
			}
		}
	}
	return "", fmt.Errorf("failed to smaple operation")
}

func runRandomizedOperations(b *testing.B, s store, totalOpsCount int, p *percentages) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < totalOpsCount; j++ {
			b.StopTimer()
			op, err := sampleOperation(p)
			require.NoError(b, err)
			b.StartTimer()

			switch op {
			case "Has":
				s.Has(randBytes(12))
			case "Get":
				s.Get(randBytes(12))
			case "Set":
				s.Set(randBytes(12), randBytes(50))
			case "Delete":
				s.Delete(randBytes(12))
			}
			if j%200 == 0 || j == totalOpsCount-1 {
				s.Commit()
			}
		}
	}
}

func prepareValues() [][]byte {
	var data [][]byte
	for i := 0; i < 5000; i++ {
		data = append(data, randBytes(50))
	}
	return data
}

func createKey(i int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(math.Sin(float64(i))*100000))
	return b
}

func runDeterministicOperations(b *testing.B, s store, values [][]byte, c *counts) {
	counts := []int{c.has, c.get, c.set, c.delete}
	sort.Ints(counts)
	step := counts[len(counts)-1]
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		idx := i * step

		b.StopTimer()
		if idx >= len(values) {
			for j := len(values); j < (idx + step); j++ {
				values = append(values, randBytes(50))
			}
		}

		b.StartTimer()
		for j := 0; j < c.set; j++ {
			key := createKey(idx + j)
			s.Set(key, values[idx+j])
		}
		for j := 0; j < c.has; j++ {
			key := createKey(idx + j)
			s.Has(key)
		}
		for j := 0; j < c.get; j++ {
			key := createKey(idx + j)
			s.Get(key)
		}
		for j := 0; j < c.delete; j++ {
			key := createKey(idx + j)
			s.Delete(key)
		}
		s.Commit()
	}
}

type newStore func(string, dbm.BackendType, string, int) (store, error)

func newStoreV1(dbName string, dbType dbm.BackendType, dir string, cacheSize int) (store, error) {
	db, err := dbm.NewDB(dbName, dbType, dir)
	if err != nil {
		return (*storev1.Store)(nil), err
	}
	tree, err := iavl.NewMutableTree(db, cacheSize)
	if err != nil {
		return (*storev1.Store)(nil), err
	}
	s := storev1.UnsafeNewStore(tree)
	return s, nil
}

func newStoreV2(_ string, dbType dbm.BackendType, dir string, _ int) (store, error) {
	switch dbType {
	case dbm.RocksDBBackend:
		db, err := rocksdb.NewDB(dir)
		if err != nil {
			return (*storev2.Store)(nil), err
		}
		s, err := storev2.NewStore(db, storev2.DefaultStoreConfig)
		if err != nil {
			return (*storev2.Store)(nil), err
		}
		return s, nil
	case dbm.BadgerDBBackend:
		db, err := badgerdb.NewDB(dir)
		if err != nil {
			return (*storev2.Store)(nil), err
		}
		s, err := storev2.NewStore(db, storev2.DefaultStoreConfig)
		if err != nil {
			return (*storev2.Store)(nil), err
		}
		return s, nil
	default:
		return (*storev2.Store)(nil), fmt.Errorf("not supported backend for store v2")
	}
}

func runSuit(b *testing.B, newStore newStore, dbBackendTypes []dbm.BackendType, dir string) {
	// run randomized operations subbenchmarks for various scenarios
	sampledPercentages := generateSampledPercentages()
	benchmarks := generateBenchmarks(dbBackendTypes, sampledPercentages, nil)
	for _, bm := range benchmarks {
		d := filepath.Join(dir, bm.name, bm.name+".db")
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			panic(err)
		}
		s, err := newStore(bm.name, bm.dbType, d, cacheSize)
		require.NoError(b, err)
		b.Run(bm.name, func(sub *testing.B) {
			runRandomizedOperations(sub, s, 1000, bm.percentages)
		})
	}

	// run deterministic operations subbenchmarks for various scenarios
	c := &counts{has: 5, get: 20, set: 5, delete: 1}
	sampledCounts := []*counts{c}
	benchmarks = generateBenchmarks(dbBackendTypes, nil, sampledCounts)
	values := prepareValues()
	for _, bm := range benchmarks {
		d := filepath.Join(dir, bm.name, bm.name+".db")
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			panic(err)
		}
		s, err := newStore(bm.name, bm.dbType, d, cacheSize)
		require.NoError(b, err)
		b.Run(bm.name, func(sub *testing.B) {
			runDeterministicOperations(sub, s, values, bm.counts)
		})
	}
}

func BenchmarkLoadStoreV1(b *testing.B) {
	dbBackendTypes := []dbm.BackendType{dbm.GoLevelDBBackend, dbm.RocksDBBackend, dbm.BadgerDBBackend}
	// dbBackendTypes := []dbm.BackendType{dbm.RocksDBBackend, dbm.BadgerDBBackend}
	runSuit(b, newStoreV1, dbBackendTypes, "testdbs/v1")
}

func BenchmarkLoadStoreV2(b *testing.B) {
	dbBackendTypes := []dbm.BackendType{dbm.RocksDBBackend, dbm.BadgerDBBackend}
	runSuit(b, newStoreV2, dbBackendTypes, "testdbs/v2")
}
