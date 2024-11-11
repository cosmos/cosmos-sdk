package gen

import (
	"fmt"
	"iter"
	"math/rand/v2"

	"github.com/cespare/xxhash/v2"

	module "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/x/benchmark"
)

type Options struct {
	*module.GeneratorParams
	// HomeDir is for reading/writing state
	HomeDir string

	InsertWeight float64
	UpdateWeight float64
	GetWeight    float64
	DeleteWeight float64
}

type State struct {
	Keys map[uint64]map[uint64]bool
}

type Generator struct {
	Options

	src    rand.Source
	rand   *rand.Rand
	state  *State
	window uint64
}

func NewGenerator(opts Options) *Generator {
	g := &Generator{
		Options: opts,
		src:     rand.NewPCG(opts.Seed, opts.Seed>>32),
		window:  100_000,
	}
	g.rand = rand.New(g.src)

	// sync state to genesis seed
	// TODO: load state from disk if present
	g.state = &State{Keys: map[uint64]map[uint64]bool{}}
	if g.GeneratorParams != nil {
		for kv := range g.GenesisSet() {
			if _, ok := g.state.Keys[kv.StoreKey]; !ok {
				g.state.Keys[kv.StoreKey] = map[uint64]bool{}
			}
			g.state.Keys[kv.StoreKey][kv.Key[0]] = true
		}
	}

	return g
}

type KV struct {
	StoreKey uint64
	Key      [2]uint64
	Value    [2]uint64
}

func (g *Generator) fetchKey(bucket uint64) (uint64, error) {
	bucketLen := uint64(len(g.state.Keys[bucket]))
	if bucketLen == 0 {
		return 0, fmt.Errorf("no keys in bucket %d", bucket)
	}
	return g.rand.Uint64N(bucketLen), nil
}

func (g *Generator) GenesisSet() iter.Seq[*KV] {
	return func(yield func(*KV) bool) {
		for range g.GenesisCount {
			seed := g.rand.Uint64()
			if !yield(&KV{
				StoreKey: g.UintN(g.BucketCount),
				Key:      [2]uint64{seed, g.getLength(g.KeyMean, g.KeyStdDev)},
				Value:    [2]uint64{seed, g.getLength(g.ValueMean, g.ValueStdDev)},
			}) {
				return
			}
		}
	}
}

func (g *Generator) Next() (uint64, *benchmark.Op, error) {
	if g.InsertWeight+g.UpdateWeight+g.GetWeight+g.DeleteWeight != 1 {
		return 0, nil, fmt.Errorf("probabilities must sum to 1")
	}

	var err error
	x := g.rand.Float64()
	bucket := g.UintN(g.BucketCount)
	op := &benchmark.Op{
		KeyLength: g.getLength(g.KeyMean, g.KeyStdDev),
		Exists:    true,
	}

	switch {
	case x < g.InsertWeight:
		// insert
		op.Seed = g.rand.Uint64()
		op.ValueLength = g.getLength(g.ValueMean, g.ValueStdDev)
		op.Exists = false
	case x < g.InsertWeight+g.UpdateWeight:
		// update
		op.Seed, err = g.fetchKey(bucket)
		if err != nil {
			return 0, nil, err
		}
		op.ValueLength = g.getLength(g.ValueMean, g.ValueStdDev)
	case x < g.InsertWeight+g.UpdateWeight+g.GetWeight:
		// get
		op.Seed, err = g.fetchKey(bucket)
		if err != nil {
			return 0, nil, err
		}
	default:
		// delete
		op.Seed, err = g.fetchKey(bucket)
		if err != nil {
			return 0, nil, err
		}
		op.Delete = true
	}

	return bucket, op, nil
}

func (g *Generator) NormUint64(mean, stdDev uint64) uint64 {
	return uint64(g.rand.NormFloat64()*float64(stdDev) + float64(mean))
}

func (g *Generator) getLength(mean, stdDev uint64) uint64 {
	length := g.NormUint64(mean, stdDev)
	if length == 0 {
		length = 1
	}
	return length
}

func (g *Generator) UintN(n uint64) uint64 {
	return g.rand.Uint64N(n)
}

func encodeUint64(x uint64) []byte {
	var b [8]byte
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	b[2] = byte(x >> 16)
	b[3] = byte(x >> 24)
	b[4] = byte(x >> 32)
	b[5] = byte(x >> 40)
	b[6] = byte(x >> 48)
	b[7] = byte(x >> 56)
	return b[:]
}

const maxStoreKeyGenIterations = 100

func StoreKeys(prefix string, seed, count uint64) ([]string, error) {
	r := rand.New(rand.NewPCG(seed, seed>>32))
	keys := make([]string, count)
	seen := make(map[string]struct{})

	var i, j uint64
	for i < count {
		if j > maxStoreKeyGenIterations {
			return nil, fmt.Errorf("failed to generate %d unique store keys", count)
		}
		sk := fmt.Sprintf("%s_%x", prefix, Bytes(r.Uint64(), 8))
		if _, ok := seen[sk]; ok {
			j++
			continue
		}
		keys[i] = sk
		seen[sk] = struct{}{}
		i++
	}
	return keys, nil
}

func Bytes(seed, length uint64) []byte {
	b := make([]byte, length)
	rounds := length / 8
	remainder := length % 8
	var h uint64
	for i := uint64(0); i < rounds; i++ {
		h = xxhash.Sum64(encodeUint64(seed + i))
		for j := uint64(0); j < 8; j++ {
			b[i*8+j] = byte(h >> (8 * j))
		}
	}
	if remainder > 0 {
		h = xxhash.Sum64(encodeUint64(seed + rounds))
		for j := uint64(0); j < remainder; j++ {
			b[rounds*8+j] = byte(h >> (8 * j))
		}
	}
	return b
}
