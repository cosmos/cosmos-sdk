package gen

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"math/rand/v2"
	"os"

	"github.com/cespare/xxhash/v2"

	module "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/tools/benchmark"
)

// Options is the configuration for the generator.
type Options struct {
	*module.GeneratorParams
	// HomeDir is for reading/writing state
	HomeDir string

	InsertWeight float64
	UpdateWeight float64
	GetWeight    float64
	DeleteWeight float64
}

// State is the state of the generator.
// It can be marshaled and unmarshaled to/from a binary format.
type State struct {
	Src interface {
		rand.Source
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
	}
	Keys [][]Payload
}

// Marshal writes the state to w.
func (s *State) Marshal(w io.Writer) error {
	srcBz, err := s.Src.MarshalBinary()
	if err != nil {
		return err
	}
	var n int
	n, err = w.Write(srcBz)
	if err != nil {
		return err
	}
	if n != 20 {
		return fmt.Errorf("expected 20 bytes, got %d", n)
	}
	if err = binary.Write(w, binary.LittleEndian, uint64(len(s.Keys))); err != nil {
		return err
	}
	for _, bucket := range s.Keys {
		if err = binary.Write(w, binary.LittleEndian, uint64(len(bucket))); err != nil {
			return err
		}
		for _, key := range bucket {
			if err = binary.Write(w, binary.LittleEndian, key); err != nil {
				return err
			}
		}
	}
	return nil
}

// Unmarshal reads the state from r.
func (s *State) Unmarshal(r io.Reader) error {
	srcBz := make([]byte, 20)
	if _, err := r.Read(srcBz); err != nil {
		return err
	}
	s.Src = rand.NewPCG(0, 0)
	if err := s.Src.UnmarshalBinary(srcBz); err != nil {
		return err
	}

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

// Generator generates operations for a benchmark transaction.
// The generator is stateful, keeping track of which keys have been inserted
// so that meaningful gets and deletes can be generated.
type Generator struct {
	Options

	rand  *rand.Rand
	state *State
}

type opt func(*Generator)

// NewGenerator creates a new generator with the given options.
func NewGenerator(opts Options, f ...opt) *Generator {
	g := &Generator{
		Options: opts,
		state: &State{
			Src: rand.NewPCG(opts.Seed, opts.Seed>>32),
		},
	}
	g.rand = rand.New(g.state.Src)
	for _, fn := range f {
		fn(g)
	}
	return g
}

// WithGenesis sets the generator state to the genesis seed.
// When the generator is created, it will sync to genesis state.
// The benchmark client needs to do this so that it can generate meaningful tx operations.
func WithGenesis() func(*Generator) {
	return func(g *Generator) {
		// sync state to genesis seed
		g.state.Keys = make([][]Payload, g.BucketCount)
		if g.GeneratorParams != nil {
			for kv := range g.GenesisSet() {
				g.state.Keys[kv.StoreKey] = append(g.state.Keys[kv.StoreKey], kv.Key)
			}
		}
	}
}

// WithSeed sets the seed for the generator.
func WithSeed(seed uint64) func(*Generator) {
	return func(g *Generator) {
		g.state.Src = rand.NewPCG(seed, seed>>32)
		g.rand = rand.New(g.state.Src)
	}
}

// Load loads the generator state from disk.
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

// Payload is a 2-tuple of seed and length.
// A seed is uint64 which is used to generate a byte slice of size length.
type Payload [2]uint64

// Seed returns the seed in the payload.
func (p Payload) Seed() uint64 {
	return p[0]
}

// Length returns the length in the payload.
func (p Payload) Length() uint64 {
	return p[1]
}

// Bytes returns the byte slice generated from the seed and length.
// The underlying byte slice is deterministically generated using the (very fast) xxhash algorithm.
func (p Payload) Bytes() []byte {
	return Bytes(p.Seed(), p.Length())
}

func (p Payload) String() string {
	return fmt.Sprintf("(%d, %d)", p.Seed(), p.Length())
}

func NewPayload(seed, length uint64) Payload {
	return Payload{seed, length}
}

// KV is a key-value pair with a store key.
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

// GenesisSet returns a sequence of key-value pairs for the genesis state.
// It is called by the server during InitGenesis to generate and set the initial state.
// The client uses WithGenesis to sync to the genesis state.
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

// Next generates the next benchmark operation.
// The operation is one of insert, update, get, or delete.
// The tx client calls this function to deterministically generate the next operation.
func (g *Generator) Next() (uint64, *benchmark.Op, error) {
	if g.InsertWeight+g.UpdateWeight+g.GetWeight+g.DeleteWeight != 1 {
		return 0, nil, fmt.Errorf("weights must sum to 1")
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

// NormUint64 returns a random uint64 with a normal distribution.
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

// UintN returns a random uint64 in the range [0, n).
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

// StoreKeys deterministically generates a set of unique store keys from seed.
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
		j++
	}
	return keys, nil
}

// Bytes generates a byte slice of length length from seed.
// The byte slice is deterministically generated using the (very fast) xxhash algorithm.
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
