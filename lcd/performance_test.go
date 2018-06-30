package lcd

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

func TestMemStoreProvidergetByHeightBinaryAndLinearSameResult(t *testing.T) {
	p := NewMemStoreProvider().(*memStoreProvider)

	// Store a bunch of commits at specific heights
	// and then ensure that:
	//  * getByHeightLinearSearch
	//  * getByHeightBinarySearch
	// both return the exact same result

	// 1. Non-existent height commits
	nonExistent := []int64{-1000, -1, 0, 1, 10, 11, 17, 31, 67, 1000, 1e9}
	ensureNonExistentCommitsAtHeight(t, "getByHeightLinearSearch", p.getByHeightLinearSearch, nonExistent)
	ensureNonExistentCommitsAtHeight(t, "getByHeightBinarySearch", p.getByHeightBinarySearch, nonExistent)

	// 2. Save some known height commits
	knownHeights := []int64{0, 1, 7, 9, 12, 13, 18, 44, 23, 16, 1024, 100, 199, 1e9}
	createAndStoreCommits(t, p, knownHeights)

	// 3. Now check if those heights are retrieved
	ensureExistentCommitsAtHeight(t, "getByHeightLinearSearch", p.getByHeightLinearSearch, knownHeights)
	ensureExistentCommitsAtHeight(t, "getByHeightBinarySearch", p.getByHeightBinarySearch, knownHeights)

	// 4. And now for the height probing to ensure that any height
	// requested returns a fullCommit of height <= requestedHeight.
	comparegetByHeightAlgorithms(t, p, 0, 0)
	comparegetByHeightAlgorithms(t, p, 1, 1)
	comparegetByHeightAlgorithms(t, p, 2, 1)
	comparegetByHeightAlgorithms(t, p, 5, 1)
	comparegetByHeightAlgorithms(t, p, 7, 7)
	comparegetByHeightAlgorithms(t, p, 10, 9)
	comparegetByHeightAlgorithms(t, p, 12, 12)
	comparegetByHeightAlgorithms(t, p, 14, 13)
	comparegetByHeightAlgorithms(t, p, 19, 18)
	comparegetByHeightAlgorithms(t, p, 43, 23)
	comparegetByHeightAlgorithms(t, p, 45, 44)
	comparegetByHeightAlgorithms(t, p, 1025, 1024)
	comparegetByHeightAlgorithms(t, p, 101, 100)
	comparegetByHeightAlgorithms(t, p, 1e3, 199)
	comparegetByHeightAlgorithms(t, p, 1e4, 1024)
	comparegetByHeightAlgorithms(t, p, 1e9, 1e9)
	comparegetByHeightAlgorithms(t, p, 1e9+1, 1e9)
}

func createAndStoreCommits(t *testing.T, p Provider, heights []int64) {
	chainID := "cache-best-height-binary-and-linear"
	appHash := []byte("0xdeadbeef")
	keys := GenValKeys(len(heights) / 2)

	for _, h := range heights {
		vals := keys.ToValidators(10, int64(len(heights)/2))
		fc := keys.GenFullCommit(chainID, h, nil, vals, appHash, []byte("params"), []byte("results"), 0, 5)
		err := p.StoreCommit(fc)
		require.NoError(t, err, "StoreCommit height=%d", h)
	}
}

func comparegetByHeightAlgorithms(t *testing.T, p *memStoreProvider, ask, expect int64) {
	algos := map[string]func(int64) (FullCommit, error){
		"getHeightByLinearSearch": p.getByHeightLinearSearch,
		"getHeightByBinarySearch": p.getByHeightBinarySearch,
	}

	for algo, fn := range algos {
		fc, err := fn(ask)
		// t.Logf("%s got=%v want=%d", algo, expect, fc.Height())
		require.Nil(t, err, "%s: %+v", algo, err)
		if assert.Equal(t, expect, fc.Height()) {
			err = p.StoreCommit(fc)
			require.Nil(t, err, "%s: %+v", algo, err)
		}
	}
}

var blankFullCommit FullCommit

func ensureNonExistentCommitsAtHeight(t *testing.T, prefix string, fn func(int64) (FullCommit, error), data []int64) {
	for i, qh := range data {
		fc, err := fn(qh)
		assert.NotNil(t, err, "#%d: %s: height=%d should return non-nil error", i, prefix, qh)
		assert.Equal(t, fc, blankFullCommit, "#%d: %s: height=%d\ngot =%+v\nwant=%+v", i, prefix, qh, fc, blankFullCommit)
	}
}

func ensureExistentCommitsAtHeight(t *testing.T, prefix string, fn func(int64) (FullCommit, error), data []int64) {
	for i, qh := range data {
		fc, err := fn(qh)
		assert.Nil(t, err, "#%d: %s: height=%d should not return an error: %v", i, prefix, qh, err)
		assert.NotEqual(t, fc, blankFullCommit, "#%d: %s: height=%d got a blankCommit", i, prefix, qh)
	}
}

func BenchmarkGenCommit20(b *testing.B) {
	keys := GenValKeys(20)
	benchmarkGenCommit(b, keys)
}

func BenchmarkGenCommit100(b *testing.B) {
	keys := GenValKeys(100)
	benchmarkGenCommit(b, keys)
}

func BenchmarkGenCommitSec20(b *testing.B) {
	keys := GenSecpValKeys(20)
	benchmarkGenCommit(b, keys)
}

func BenchmarkGenCommitSec100(b *testing.B) {
	keys := GenSecpValKeys(100)
	benchmarkGenCommit(b, keys)
}

func benchmarkGenCommit(b *testing.B, keys ValKeys) {
	chainID := fmt.Sprintf("bench-%d", len(keys))
	vals := keys.ToValidators(20, 10)
	for i := 0; i < b.N; i++ {
		h := int64(1 + i)
		appHash := []byte(fmt.Sprintf("h=%d", h))
		resHash := []byte(fmt.Sprintf("res=%d", h))
		keys.GenCommit(chainID, h, nil, vals, appHash, []byte("params"), resHash, 0, len(keys))
	}
}

// this benchmarks generating one key
func BenchmarkGenValKeys(b *testing.B) {
	keys := GenValKeys(20)
	for i := 0; i < b.N; i++ {
		keys = keys.Extend(1)
	}
}

// this benchmarks generating one key
func BenchmarkGenSecpValKeys(b *testing.B) {
	keys := GenSecpValKeys(20)
	for i := 0; i < b.N; i++ {
		keys = keys.Extend(1)
	}
}

func BenchmarkToValidators20(b *testing.B) {
	benchmarkToValidators(b, 20)
}

func BenchmarkToValidators100(b *testing.B) {
	benchmarkToValidators(b, 100)
}

// this benchmarks constructing the validator set (.PubKey() * nodes)
func benchmarkToValidators(b *testing.B, nodes int) {
	keys := GenValKeys(nodes)
	for i := 1; i <= b.N; i++ {
		keys.ToValidators(int64(2*i), int64(i))
	}
}

func BenchmarkToValidatorsSec100(b *testing.B) {
	benchmarkToValidatorsSec(b, 100)
}

// this benchmarks constructing the validator set (.PubKey() * nodes)
func benchmarkToValidatorsSec(b *testing.B, nodes int) {
	keys := GenSecpValKeys(nodes)
	for i := 1; i <= b.N; i++ {
		keys.ToValidators(int64(2*i), int64(i))
	}
}

func BenchmarkCertifyCommit20(b *testing.B) {
	keys := GenValKeys(20)
	benchmarkCertifyCommit(b, keys)
}

func BenchmarkCertifyCommit100(b *testing.B) {
	keys := GenValKeys(100)
	benchmarkCertifyCommit(b, keys)
}

func BenchmarkCertifyCommitSec20(b *testing.B) {
	keys := GenSecpValKeys(20)
	benchmarkCertifyCommit(b, keys)
}

func BenchmarkCertifyCommitSec100(b *testing.B) {
	keys := GenSecpValKeys(100)
	benchmarkCertifyCommit(b, keys)
}

func benchmarkCertifyCommit(b *testing.B, keys ValKeys) {
	chainID := "bench-certify"
	vals := keys.ToValidators(20, 10)
	cert := NewStaticCertifier(chainID, vals)
	check := keys.GenCommit(chainID, 123, nil, vals, []byte("foo"), []byte("params"), []byte("res"), 0, len(keys))
	for i := 0; i < b.N; i++ {
		err := cert.Certify(check)
		if err != nil {
			panic(err)
		}
	}

}

type algo bool

const (
	linearSearch = true
	binarySearch = false
)

// Lazy load the commits
var fcs5, fcs50, fcs100, fcs500, fcs1000 []FullCommit
var h5, h50, h100, h500, h1000 []int64
var commitsOnce sync.Once

func lazyGenerateFullCommits(b *testing.B) {
	b.Logf("Generating FullCommits")
	commitsOnce.Do(func() {
		fcs5, h5 = genFullCommits(nil, nil, 5)
		b.Logf("Generated 5 FullCommits")
		fcs50, h50 = genFullCommits(fcs5, h5, 50)
		b.Logf("Generated 50 FullCommits")
		fcs100, h100 = genFullCommits(fcs50, h50, 100)
		b.Logf("Generated 100 FullCommits")
		fcs500, h500 = genFullCommits(fcs100, h100, 500)
		b.Logf("Generated 500 FullCommits")
		fcs1000, h1000 = genFullCommits(fcs500, h500, 1000)
		b.Logf("Generated 1000 FullCommits")
	})
}

func BenchmarkMemStoreProviderGetByHeightLinearSearch5(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs5, h5, linearSearch)
}

func BenchmarkMemStoreProviderGetByHeightLinearSearch50(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs50, h50, linearSearch)
}

func BenchmarkMemStoreProviderGetByHeightLinearSearch100(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs100, h100, linearSearch)
}

func BenchmarkMemStoreProviderGetByHeightLinearSearch500(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs500, h500, linearSearch)
}

func BenchmarkMemStoreProviderGetByHeightLinearSearch1000(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs1000, h1000, linearSearch)
}

func BenchmarkMemStoreProviderGetByHeightBinarySearch5(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs5, h5, binarySearch)
}

func BenchmarkMemStoreProviderGetByHeightBinarySearch50(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs50, h50, binarySearch)
}

func BenchmarkMemStoreProviderGetByHeightBinarySearch100(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs100, h100, binarySearch)
}

func BenchmarkMemStoreProviderGetByHeightBinarySearch500(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs500, h500, binarySearch)
}

func BenchmarkMemStoreProviderGetByHeightBinarySearch1000(b *testing.B) {
	benchmarkMemStoreProvidergetByHeight(b, fcs1000, h1000, binarySearch)
}

var rng = rand.New(rand.NewSource(10))

func benchmarkMemStoreProvidergetByHeight(b *testing.B, fcs []FullCommit, fHeights []int64, algo algo) {
	lazyGenerateFullCommits(b)

	b.StopTimer()
	mp := NewMemStoreProvider()
	for i, fc := range fcs {
		if err := mp.StoreCommit(fc); err != nil {
			b.Fatalf("FullCommit #%d: err: %v", i, err)
		}
	}
	qHeights := make([]int64, len(fHeights))
	copy(qHeights, fHeights)
	// Append some non-existent heights to trigger the worst cases.
	qHeights = append(qHeights, 19, -100, -10000, 1e7, -17, 31, -1e9)

	memP := mp.(*memStoreProvider)
	searchFn := memP.getByHeightLinearSearch
	if algo == binarySearch { // nolint
		searchFn = memP.getByHeightBinarySearch
	}

	hPerm := rng.Perm(len(qHeights))
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, j := range hPerm {
			h := qHeights[j]
			if _, err := searchFn(h); err != nil {
			}
		}
	}
	b.ReportAllocs()
}

func genFullCommits(prevFC []FullCommit, prevH []int64, want int) ([]FullCommit, []int64) {
	fcs := make([]FullCommit, len(prevFC))
	copy(fcs, prevFC)
	heights := make([]int64, len(prevH))
	copy(heights, prevH)

	appHash := []byte("benchmarks")
	chainID := "benchmarks-gen-full-commits"
	n := want
	keys := GenValKeys(2 + (n / 3))
	for i := 0; i < n; i++ {
		vals := keys.ToValidators(10, int64(n/2))
		h := int64(20 + 10*i)
		fcs = append(fcs, keys.GenFullCommit(chainID, h, nil, vals, appHash, []byte("params"), []byte("results"), 0, 5))
		heights = append(heights, h)
	}
	return fcs, heights
}

func TestMemStoreProviderLatestCommitAlwaysUsesSorted(t *testing.T) {
	p := NewMemStoreProvider().(*memStoreProvider)
	// 1. With no commits yet stored, it should return ErrCommitNotFound
	got, err := p.LatestCommit()
	require.Equal(t, err.Error(), lcdErr.ErrCommitNotFound().Error(), "should return ErrCommitNotFound()")
	require.Equal(t, got, blankFullCommit, "With no fullcommits, it should return a blank FullCommit")

	// 2. Generate some full commits now and we'll add them unsorted.
	genAndStoreCommitsOfHeight(t, p, 27, 100, 1, 12, 1000, 17, 91)
	fc, err := p.LatestCommit()
	require.Nil(t, err, "with commits saved no error expected")
	require.NotEqual(t, fc, blankFullCommit, "with commits saved no blank FullCommit")
	require.Equal(t, fc.Height(), int64(1000), "the latest commit i.e. the largest expected")
}

func genAndStoreCommitsOfHeight(t *testing.T, p Provider, heights ...int64) {
	n := len(heights)
	appHash := []byte("tests")
	chainID := "tests-gen-full-commits"
	keys := GenValKeys(2 + (n / 3))
	for i := 0; i < n; i++ {
		h := heights[i]
		vals := keys.ToValidators(10, int64(n/2))
		fc := keys.GenFullCommit(chainID, h, nil, vals, appHash, []byte("params"), []byte("results"), 0, 5)
		err := p.StoreCommit(fc)
		require.NoError(t, err, "StoreCommit height=%d", h)
	}
}
