package gen

import (
	"math/rand/v2"

	"github.com/cespare/xxhash/v2"

	"cosmossdk.io/x/benchmark"
)

type Options struct {
	Seed uint64

	KeyMean     uint64
	KeyStdDev   uint64
	ValueMean   uint64
	ValueStdDev uint64

	DeleteFraction float64
}

type State struct {
	Keys map[int]map[int]struct{}
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
		rand:    rand.New(rand.NewPCG(opts.Seed, opts.Seed>>32)),
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

func (g *Generator) NormUint64(mean, stdDev uint64) uint64 {
	return uint64(g.rand.NormFloat64()*float64(stdDev) + float64(mean))
}

func (g *Generator) Key() []byte {
	length := g.NormUint64(g.KeyMean, g.KeyStdDev)
	seed := g.rand.Uint64()
	return g.Bytes(seed, length)
}

func (g *Generator) Value() []byte {
	length := g.NormUint64(g.ValueMean, g.ValueStdDev)
	seed := g.rand.Uint64()
	return g.Bytes(seed, length)
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
