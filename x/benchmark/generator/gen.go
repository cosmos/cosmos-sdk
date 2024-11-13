package gen

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"math/rand/v2"
	"os"

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
	Keys [][]Payload
}

func (s *State) Marshal(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, uint64(len(s.Keys))); err != nil {
		return err
	}
	for _, bucket := range s.Keys {
		if err := binary.Write(w, binary.LittleEndian, uint64(len(bucket))); err != nil {
			return err
		}
		for _, key := range bucket {
			if err := binary.Write(w, binary.LittleEndian, key); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *State) Unmarshal(r io.Reader) error {
	var n uint64
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return err
	}
	s.Keys = make([][]Payload, n)
	for i := uint64(0); i < n; i++ {
		var m uint64
		if err := binary.Read(r, binary.LittleEndian, &m); err != nil {
			return err
		}
		s.Keys[i] = make([]Payload, m)
		for j := uint64(0); j < m; j++ {
			if err := binary.Read(r, binary.LittleEndian, &s.Keys[i][j]); err != nil {
				return err
			}
		}
	}
	return nil
}

type Generator struct {
	Options

	src   rand.Source
	rand  *rand.Rand
	state *State
}

type opt func(*Generator)

func NewGenerator(opts Options, f ...opt) *Generator {
	g := &Generator{
		Options: opts,
		src:     rand.NewPCG(opts.Seed, opts.Seed>>32),
	}
	g.rand = rand.New(g.src)
	for _, fn := range f {
		fn(g)
	}
	return g
}

func (g *Generator) Load() error {
	f := fmt.Sprintf("%s/data/generator_state.bin", g.HomeDir)
	r, err := os.Open(f)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return g.state.Unmarshal(r)
}

func WithGenesis() func(*Generator) {
	return func(g *Generator) {
		// sync state to genesis seed
		g.state = &State{Keys: make([][]Payload, g.BucketCount)}
		if g.GeneratorParams != nil {
			for kv := range g.GenesisSet() {
				g.state.Keys[kv.StoreKey] = append(g.state.Keys[kv.StoreKey], kv.Key)
			}
		}
	}
}

func WithSeed(seed uint64) func(*Generator) {
	return func(g *Generator) {
		g.src = rand.NewPCG(seed, seed>>32)
		g.rand = rand.New(g.src)
	}
}

type Payload [2]uint64

func (p Payload) Seed() uint64 {
	return p[0]
}

func (p Payload) Length() uint64 {
	return p[1]
}

func (p Payload) Bytes() []byte {
	return Bytes(p.Seed(), p.Length())
}

func (p Payload) String() string {
	return fmt.Sprintf("(%d, %d)", p.Seed(), p.Length())
}

func NewPayload(seed, length uint64) Payload {
	return Payload{seed, length}
}

type KV struct {
	StoreKey uint64
	Key      Payload
	Value    Payload
}

func (g *Generator) fetchKey(bucket uint64) (idx uint64, key Payload, err error) {
	bucketLen := uint64(len(g.state.Keys[bucket]))
	if bucketLen == 0 {
		return 0, Payload{}, fmt.Errorf("no keys in bucket %d", bucket)
	}
	idx = g.rand.Uint64N(bucketLen)
	return idx, g.state.Keys[bucket][idx], nil
}

func (g *Generator) deleteKey(bucket, idx uint64) {
	g.state.Keys[bucket] = append(g.state.Keys[bucket][:idx], g.state.Keys[bucket][idx+1:]...)
}

func (g *Generator) setKey(bucket uint64, payload Payload) {
	g.state.Keys[bucket] = append(g.state.Keys[bucket], payload)
}

func (g *Generator) GenesisSet() iter.Seq[*KV] {
	return func(yield func(*KV) bool) {
		for range g.GenesisCount {
			seed := g.rand.Uint64()
			if !yield(&KV{
				StoreKey: g.UintN(g.BucketCount),
				Key:      NewPayload(seed, g.getLength(g.KeyMean, g.KeyStdDev)),
				Value:    NewPayload(seed, g.getLength(g.ValueMean, g.ValueStdDev)),
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

	var (
		err error
		key Payload
	)
	x := g.rand.Float64()
	bucket := g.UintN(g.BucketCount)
	op := &benchmark.Op{
		Exists: true,
	}

	switch {
	case x < g.InsertWeight:
		// insert
		op.Seed = g.rand.Uint64()
		op.KeyLength = g.getLength(g.KeyMean, g.KeyStdDev)
		op.ValueLength = g.getLength(g.ValueMean, g.ValueStdDev)
		op.Exists = false
		g.setKey(bucket, NewPayload(op.Seed, op.KeyLength))
	case x < g.InsertWeight+g.UpdateWeight:
		// update
		_, key, err = g.fetchKey(bucket)
		if err != nil {
			return 0, nil, err
		}
		op.Seed = key.Seed()
		op.KeyLength = key.Length()
		op.ValueLength = g.getLength(g.ValueMean, g.ValueStdDev)
	case x < g.InsertWeight+g.UpdateWeight+g.GetWeight:
		// get
		_, key, err = g.fetchKey(bucket)
		if err != nil {
			return 0, nil, err
		}
		op.Seed = key.Seed()
		op.KeyLength = key.Length()
	default:
		// delete
		var idx uint64
		idx, key, err = g.fetchKey(bucket)
		if err != nil {
			return 0, nil, err
		}
		op.Delete = true
		op.Seed = key.Seed()
		op.KeyLength = key.Length()
		g.deleteKey(bucket, idx)
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

func (g *Generator) Close() error {
	f := fmt.Sprintf("%s/data/generator_state.bin", g.HomeDir)
	w, err := os.Create(f)
	if err != nil {
		return err
	}
	return g.state.Marshal(w)
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
