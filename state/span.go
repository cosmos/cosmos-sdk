package state

import wire "github.com/tendermint/go-wire"

var (
	keys = []byte("keys")
	// uses dataKey from queue.go to prefix data
)

// Span holds a number of different keys in a large range and allows
// use to make some basic range queries, like highest between, lowest between...
// All items are added with an index
//
// This becomes horribly inefficent as len(keys) => 1000+, but by then
// hopefully we have access to the iavl tree to do this well
//
// TODO: doesn't handle deleting....
type Span struct {
	store KVStore
	// keys is sorted ascending and cannot contain duplicates
	keys []uint64
}

// NewSpan loads or initializes a span of keys
func NewSpan(store KVStore) *Span {
	s := &Span{store: store}
	s.loadKeys()
	return s
}

// Set puts a value at a given height
func (s *Span) Set(h uint64, value []byte) {
	key := makeKey(h)
	s.store.Set(key, value)
	s.addKey(h)
	s.storeKeys()
}

// Get returns the element at h if it exists
func (s *Span) Get(h uint64) []byte {
	key := makeKey(h)
	return s.store.Get(key)
}

// Bottom returns the lowest element in the Span, along with its index
func (s *Span) Bottom() ([]byte, uint64) {
	if len(s.keys) == 0 {
		return nil, 0
	}
	h := s.keys[0]
	return s.Get(h), h
}

// Top returns the highest element in the Span, along with its index
func (s *Span) Top() ([]byte, uint64) {
	l := len(s.keys)
	if l == 0 {
		return nil, 0
	}
	h := s.keys[l-1]
	return s.Get(h), h
}

// GTE returns the lowest element in the Span that is >= h, along with its index
func (s *Span) GTE(h uint64) ([]byte, uint64) {
	for _, k := range s.keys {
		if k >= h {
			return s.Get(k), k
		}
	}
	return nil, 0
}

// LTE returns the highest element in the Span that is <= h,
// along with its index
func (s *Span) LTE(h uint64) ([]byte, uint64) {
	var k uint64
	// start from the highest and go down for the first match
	for i := len(s.keys) - 1; i >= 0; i-- {
		k = s.keys[i]
		if k <= h {
			return s.Get(k), k
		}
	}
	return nil, 0
}

// addKey inserts this key, maintaining sorted order, no duplicates
func (s *Span) addKey(h uint64) {
	for i, k := range s.keys {
		// don't add duplicates
		if h == k {
			return
		}
		// insert before this key
		if h < k {
			// https://github.com/golang/go/wiki/SliceTricks
			s.keys = append(s.keys, 0)
			copy(s.keys[i+1:], s.keys[i:])
			s.keys[i] = h
			return
		}
	}
	// if it is higher than all (or empty keys), append
	s.keys = append(s.keys, h)
}

func (s *Span) loadKeys() {
	b := s.store.Get(keys)
	if b == nil {
		return
	}
	err := wire.ReadBinaryBytes(b, &s.keys)
	// hahaha... just like i love to hate :)
	if err != nil {
		panic(err)
	}
}

func (s *Span) storeKeys() {
	b := wire.BinaryBytes(s.keys)
	s.store.Set(keys, b)
}
