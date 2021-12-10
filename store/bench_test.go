package store

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	storev1 "github.com/cosmos/cosmos-sdk/store/iavl"
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

func generateRatioGrid() [][4]uint {
	var ratioGrid [][4]uint
	for a := uint(0); a < 100; a += 20 {
		for b := uint(0); b < 100; b += 20 {
			if a+b > 100 {
				break
			} else {
				for c := uint(0); c < 100; c += 10 {
					if a+b+c > 100 {
						break
					} else {
						d := 100 - a - b - c
						ratio := [4]uint{a, b, c, d}
						ratioGrid = append(ratioGrid, ratio)
					}
				}
			}
		}
	}
	return ratioGrid
}

type benchmark struct {
	name  string
	ratio [4]uint
}

func generateBenchmarks() []benchmark {
	var benchmarks []benchmark
	ratios := generateRatioGrid()
	for _, ratio := range ratios {
		name := fmt.Sprintf("%d-%d-%d-%d", ratio[0], ratio[1], ratio[2], ratio[3])
		benchmarks = append(benchmarks, benchmark{name: name, ratio: ratio})
	}
	return benchmarks
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func runLoadStoreV1(b *testing.B, operationsCount uint, ratio [4]uint) {
	db := dbm.NewMemDB()
	tree, err := iavl.NewMutableTree(db, cacheSize)
	require.NoError(b, err)
	store := storev1.UnsafeNewStore(tree)

	opLimit := map[string]uint{}
	opLimit["Has"] = operationsCount * ratio[0] / 100
	opLimit["Get"] = operationsCount * ratio[1] / 100
	opLimit["Set"] = operationsCount * ratio[2] / 100
	opLimit["Delete"] = operationsCount * ratio[3] / 100

	opCount := map[string]uint{}
	keys := []string{"Has", "Get", "Set", "Delete"}
	for j := 0; j < int(operationsCount); j++ {
		r, _ := crand.Int(crand.Reader, big.NewInt(int64(len(keys))))
		k := int(r.Uint64())
		key := keys[k]
		switch key {
		case "Has":
			store.Has(randBytes(12))
			opCount["Has"] += 1
			if opCount["Has"] >= opLimit["Has"] {
				keys = remove(keys, k)
			}
		case "Get":
			store.Get(randBytes(12))
			opCount["Get"] += 1
			if opCount["Get"] >= opLimit["Get"] {
				keys = remove(keys, k)
			}
		case "Set":
			store.Set(randBytes(12), randBytes(50))
			opCount["Set"] += 1
			if opCount["Set"] >= opLimit["Set"] {
				keys = remove(keys, k)
			}
		case "Delete":
			store.Delete(randBytes(12))
			opCount["Delete"] += 1
			if opCount["Delete"] >= opLimit["Delete"] {
				keys = remove(keys, k)
			}
		}
	}

	id := store.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = storev1.LoadStore(db, id, false, storev1.DefaultIAVLCacheSize)
		require.NoError(b, err)
	}
}

func runLoadStoreV2(b *testing.B, operationsCount uint, ratio [4]uint) {
	db := memdb.NewDB()
	store, err := storev2.NewStore(db, storev2.DefaultStoreConfig)
	require.NoError(b, err)

	opLimit := map[string]uint{}
	opLimit["Has"] = operationsCount * ratio[0] / 100
	opLimit["Get"] = operationsCount * ratio[1] / 100
	opLimit["Set"] = operationsCount * ratio[2] / 100
	opLimit["Delete"] = operationsCount * ratio[3] / 100

	opCount := map[string]uint{}
	keys := []string{"Has", "Get", "Set", "Delete"}
	for j := 0; j < int(operationsCount); j++ {
		r, _ := crand.Int(crand.Reader, big.NewInt(int64(len(keys))))
		k := int(r.Uint64())
		key := keys[k]
		switch key {
		case "Has":
			store.Has(randBytes(12))
			opCount["Has"] += 1
			if opCount["Has"] >= opLimit["Has"] {
				keys = remove(keys, k)
			}
		case "Get":
			store.Get(randBytes(12))
			opCount["Get"] += 1
			if opCount["Get"] >= opLimit["Get"] {
				keys = remove(keys, k)
			}
		case "Set":
			store.Set(randBytes(12), randBytes(50))
			opCount["Set"] += 1
			if opCount["Set"] >= opLimit["Set"] {
				keys = remove(keys, k)
			}
		case "Delete":
			store.Delete(randBytes(12))
			opCount["Delete"] += 1
			if opCount["Delete"] >= opLimit["Delete"] {
				keys = remove(keys, k)
			}
		}
	}

	_ = store.Commit()
	require.NoError(b, store.Close())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store, err = storev2.NewStore(db, storev2.DefaultStoreConfig)
		require.NoError(b, err)
		require.NoError(b, store.Close())
	}
}

func BenchmarkLoadStoreV1(b *testing.B) {
	bm := benchmark{name: "test", ratio: [4]uint{5, 65, 30, 0}}
	b.Run(bm.name, func(sub *testing.B) {
		runLoadStoreV1(sub, 10000, bm.ratio)
	})
}

func BenchmarkLoadStoreV2(b *testing.B) {
	bm := benchmark{name: "test", ratio: [4]uint{5, 65, 30, 0}}
	b.Run(bm.name, func(sub *testing.B) {
		runLoadStoreV2(sub, 10000, bm.ratio)
	})
}

func BenchmarkLoadStore(b *testing.B) {
	benchmarks := generateBenchmarks()
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("%s-%s", "v1", bm.name), func(sub *testing.B) {
			runLoadStoreV1(sub, 10000, bm.ratio)
		})
		b.Run(fmt.Sprintf("%s-%s", "v2", bm.name), func(sub *testing.B) {
			runLoadStoreV2(sub, 10000, bm.ratio)
		})
	}
}
