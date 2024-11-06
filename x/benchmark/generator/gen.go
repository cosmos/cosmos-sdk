package generator

import (
	"math/rand"

	"github.com/cespare/xxhash/v2"

	"cosmossdk.io/x/benchmark"
)

type Options struct {
	Seed int64

	StoreKeys []string

	KeyMean     int
	KeyStdDev   int
	ValueMean   int
	ValueStdDev int

	// Inserts specifies the number of Insert operations to generate. If zero, inserts will
	// occur naturally.  If set to a non-zero value, inserts will be weighted in generation until
	// the target number of inserts is reached
	Inserts        int
	DeleteFraction float64
}

type State struct {
	Keys map[string]map[int]struct{}
}

type Generator struct {
	Options

	digest *xxhash.Digest
	rand   *rand.Rand
}

func NewGenerator(opts Options) *Generator {
	return &Generator{
		Options: opts,
		digest:  xxhash.New(),
		rand:    rand.New(rand.NewSource(opts.Seed)),
	}
}

func (g *Generator) Next() (*benchmark.Op, error) {
	return nil, nil
}

func (g *Generator) Bytes(seed, length uint64) []byte {
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

func (g *Generator) genLength() int64 {
	return 0
}

func genLength(seed int64) int64 {
	rand := rand.New(rand.NewSource(seed))
	rand.NormFloat64()
	return 0
}
