package store

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/badgerdb"
	"github.com/cosmos/cosmos-sdk/db/rocksdb"
	storev1 "github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/types"
	storev2types "github.com/cosmos/cosmos-sdk/store/v2"
	storev2 "github.com/cosmos/cosmos-sdk/store/v2/multi"
	tmdb "github.com/tendermint/tm-db"
)

var (
	cacheSize = 100
)

func randBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, _ = rand.Read(b)
	return b
}

type percentages struct {
	has    int
	get    int
	set    int
	delete int
}

type counts struct {
	has    int
	get    int
	set    int
	delete int
}

func generateSampledPercentages() []percentages {
	var sampledPercentages []percentages
	sampleX := percentages{has: 2, get: 55, set: 40, delete: 3}
	sampledPercentages = append(sampledPercentages, sampleX)
	for a := 0; a < 100; a += 20 {
		for b := 0; b <= 100-a; b += 20 {
			for c := 0; c < 100-a-b; c += 20 {
				sample := percentages{
					has:    a,
					get:    b,
					set:    c,
					delete: 100 - a - b - c,
				}
				sampledPercentages = append(sampledPercentages, sample)
			}
		}
	}
	return sampledPercentages
}

type benchmark struct {
	name        string
	percentages percentages
	dbType      tmdb.BackendType
	counts      counts
}

func generateBenchmarks(dbBackendTypes []tmdb.BackendType, sampledPercentages []percentages, sampledCounts []counts) []benchmark {
	var benchmarks []benchmark
	for _, dbType := range dbBackendTypes {
		if len(sampledPercentages) > 0 {
			for _, p := range sampledPercentages {
				name := fmt.Sprintf("r-%s-%d-%d-%d-%d", dbType, p.has, p.get, p.set, p.delete)
				benchmarks = append(benchmarks, benchmark{name: name, percentages: p, dbType: dbType, counts: counts{}})
			}
		} else if len(sampledCounts) > 0 {
			for _, c := range sampledCounts {
				name := fmt.Sprintf("d-%s-%d-%d-%d-%d", dbType, c.has, c.get, c.set, c.delete)
				benchmarks = append(benchmarks, benchmark{name: name, percentages: percentages{}, dbType: dbType, counts: c})
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

type storeV2 struct {
	*storev2.Store
	storev2types.KVStore
}

func sampleOperation(p percentages) string {
	ops := []string{"Has", "Get", "Set", "Delete"}
	thresholds := []int{p.has, p.has + p.get, p.has + p.get + p.set}
	r := rand.Intn(100)
	for i := 0; i < len(thresholds); i++ {
		if r < thresholds[i] {
			return ops[i]
		}
	}
	return ops[3]
}

func runRandomizedOperations(b *testing.B, s store, totalOpsCount int, p percentages) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < totalOpsCount; j++ {
			b.StopTimer()
			op := sampleOperation(p)
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

func runDeterministicOperations(b *testing.B, s store, values [][]byte, c counts) {
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

func newDB(version int, dbName string, dbType tmdb.BackendType, dir string) (db interface{}, err error) {
	d := filepath.Join(dir, dbName, dbName+".db")
	err = os.MkdirAll(d, os.ModePerm)
	if err != nil {
		panic(err)
	}

	if version == 1 {
		db, err = tmdb.NewDB(dbName, dbType, d)
		if err != nil {
			return nil, err
		}
		return db, err
	}

	if version == 2 {
		switch dbType {
		case tmdb.RocksDBBackend:
			db, err = rocksdb.NewDB(d)
			if err != nil {
				return nil, err
			}
			return db, nil
		case tmdb.BadgerDBBackend:
			db, err = badgerdb.NewDB(d)
			if err != nil {
				return nil, err
			}
			return db, nil
		default:
			return nil, fmt.Errorf("not supported backend for store v2")
		}
	}

	return nil, fmt.Errorf("not supported version")
}

func newStore(version int, dbBackend interface{}, cID *types.CommitID, cacheSize int) (store, error) {
	if version == 1 {
		db, ok := dbBackend.(tmdb.DB)
		if !ok {
			return nil, fmt.Errorf("unsupported db type")
		}
		if cID == nil {
			cID = &types.CommitID{Version: 0, Hash: nil}
		}
		s, err := storev1.LoadStore(db, *cID, false, cacheSize)
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	if version == 2 {
		db, ok := dbBackend.(db.DBConnection)
		if !ok {
			return nil, fmt.Errorf("unsupported db type")
		}
		root, err := storev2.NewStore(db, storev2.DefaultStoreConfig())
		if err != nil {
			return nil, err
		}
		store := root.GetKVStore(storev2types.NewKVStoreKey("store1"))
		s := &storeV2{root, store}
		return s, nil
	}

	return nil, fmt.Errorf("unsupported version")
}

func prepareStore(s store, values [][]byte) (store, types.CommitID) {
	for i, v := range values {
		s.Set(createKey(i), v)
	}
	cID := s.Commit()
	return s, cID
}

func runSuite(b *testing.B, version int, dbBackendTypes []tmdb.BackendType, dir string) {
	// run randomized operations subbenchmarks for various scenarios
	sampledPercentages := generateSampledPercentages()
	benchmarks := generateBenchmarks(dbBackendTypes, sampledPercentages, nil)
	for _, bm := range benchmarks {
		db, err := newDB(version, bm.name, bm.dbType, dir)
		require.NoError(b, err)
		s, err := newStore(version, db, nil, cacheSize)
		require.NoError(b, err)
		b.Run(bm.name, func(sub *testing.B) {
			runRandomizedOperations(sub, s, 1000, bm.percentages)
		})
	}

	// run deterministic operations subbenchmarks for various scenarios
	c := counts{has: 5, get: 20, set: 5, delete: 1}
	sampledCounts := []counts{c}
	benchmarks = generateBenchmarks(dbBackendTypes, nil, sampledCounts)
	values := prepareValues()
	for _, bm := range benchmarks {
		db, err := newDB(version, bm.name, bm.dbType, dir)
		require.NoError(b, err)
		s, err := newStore(version, db, nil, cacheSize)
		require.NoError(b, err)
		b.Run(bm.name, func(sub *testing.B) {
			runDeterministicOperations(sub, s, values, bm.counts)
		})
	}
}

func BenchmarkLoadStoreV1(b *testing.B) {
	dbBackendTypes := []tmdb.BackendType{tmdb.GoLevelDBBackend, tmdb.RocksDBBackend, tmdb.BadgerDBBackend}
	// dbBackendTypes := []tmdb.BackendType{tmdb.RocksDBBackend, tmdb.BadgerDBBackend}
	runSuite(b, 1, dbBackendTypes, "testdbs/v1")
}

func BenchmarkLoadStoreV2(b *testing.B) {
	dbBackendTypes := []tmdb.BackendType{tmdb.RocksDBBackend, tmdb.BadgerDBBackend}
	runSuite(b, 2, dbBackendTypes, "testdbs/v2")
}
