package v2

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/rand"
)

const (
	rngMax  = 1 << 63
	rngMask = rngMax - 1
)

// RandSource defines an interface for random number sources with a method to retrieve the seed.
type RandSource interface {
	rand.Source
	GetSeed() int64
}

var (
	_ RandSource = &SeededRandomSource{}
	_ RandSource = &ByteSource{}
)

// SeededRandomSource wraps a random source with an associated seed value for reproducible random number generation.
// It implements the RandSource interface, allowing access to both the random source and seed.
type SeededRandomSource struct {
	rand.Source
	seed int64
}

// NewSeededRandSource constructor
func NewSeededRandSource(seed int64) *SeededRandomSource {
	r := new(SeededRandomSource)
	r.Seed(seed)
	return r
}

func (r *SeededRandomSource) Seed(seed int64) {
	r.seed = seed
	r.Source = rand.NewSource(seed)
}

func (r SeededRandomSource) GetSeed() int64 {
	return r.seed
}

// ByteSource offers deterministic pseudo-random numbers for math.Rand with fuzzer support.
// The 'seed' data is read in big endian to uint64. When exhausted,
// it falls back to a standard random number generator initialized with a specific 'seed' value.
type ByteSource struct {
	seed     *bytes.Reader
	fallback *rand.Rand
}

// NewByteSource creates a new ByteSource with a specified byte slice and seed. This gives a fixed sequence of pseudo-random numbers.
// Initially, it utilizes the byte slice. Once that's exhausted, it continues generating numbers using the provided seed.
func NewByteSource(fuzzSeed []byte, seed int64) *ByteSource {
	return &ByteSource{
		seed:     bytes.NewReader(fuzzSeed),
		fallback: rand.New(rand.NewSource(seed)),
	}
}

func (s *ByteSource) Uint64() uint64 {
	if s.seed.Len() < 8 {
		return s.fallback.Uint64()
	}
	var b [8]byte
	if _, err := s.seed.Read(b[:]); err != nil && err != io.EOF {
		panic(err) // Should not happen.
	}
	return binary.BigEndian.Uint64(b[:])
}

func (s *ByteSource) Int63() int64 {
	return int64(s.Uint64() & rngMask)
}

// Seed is not supported and will panic
func (s *ByteSource) Seed(seed int64) {
	panic("not supported")
}

// GetSeed is not supported and will panic
func (s ByteSource) GetSeed() int64 {
	panic("not supported")
}
