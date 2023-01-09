package collections

import (
	"fmt"
	"strings"
)

// Pair defines a key composed of two keys.
type Pair[K1, K2 any] struct {
	key1 *K1
	key2 *K2
}

// K1 returns the first part of the key.
// If not present the zero value is returned.
func (p Pair[K1, K2]) K1() (k1 K1) {
	if p.key1 == nil {
		return
	}
	return *p.key1
}

// K2 returns the second part of the key.
// If not present the zero value is returned.
func (p Pair[K1, K2]) K2() (k2 K2) {
	if p.key2 == nil {
		return
	}
	return *p.key2
}

// Join creates a new Pair instance composed of the two provided keys, in order.
func Join[K1, K2 any](key1 K1, key2 K2) Pair[K1, K2] {
	return Pair[K1, K2]{
		key1: &key1,
		key2: &key2,
	}
}

// PairPrefix creates a new Pair instance composed only of the first part of the key.
func PairPrefix[K1, K2 any](key K1) Pair[K1, K2] {
	return Pair[K1, K2]{key1: &key}
}

// PairSuffix creates a new Pair instance composed only of the second part of the key.
func PairSuffix[K1, K2 any](key K2) Pair[K1, K2] {
	return Pair[K1, K2]{key2: &key}
}

// PairKeyCodec instantiates a new KeyCodec instance that can encode the Pair, given the KeyCodec of the
// first part of the key and the KeyCodec of the second part of the key.
func PairKeyCodec[K1, K2 any](keyCodec1 KeyCodec[K1], keyCodec2 KeyCodec[K2]) KeyCodec[Pair[K1, K2]] {
	return pairKeyCodec[K1, K2]{
		keyCodec1: keyCodec1,
		keyCodec2: keyCodec2,
	}
}

type pairKeyCodec[K1, K2 any] struct {
	keyCodec1 KeyCodec[K1]
	keyCodec2 KeyCodec[K2]
}

func (p pairKeyCodec[K1, K2]) Encode(buffer []byte, pair Pair[K1, K2]) (int, error) {
	writtenTotal := 0
	if pair.key1 != nil {
		written, err := p.keyCodec1.EncodeNonTerminal(buffer, *pair.key1)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if pair.key2 != nil {
		written, err := p.keyCodec2.Encode(buffer[writtenTotal:], *pair.key2)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	return writtenTotal, nil
}

func (p pairKeyCodec[K1, K2]) Decode(buffer []byte) (int, Pair[K1, K2], error) {
	readTotal := 0
	read, key1, err := p.keyCodec1.DecodeNonTerminal(buffer)
	if err != nil {
		return 0, Pair[K1, K2]{}, err
	}
	readTotal += read
	read, key2, err := p.keyCodec2.Decode(buffer[read:])
	if err != nil {
		return 0, Pair[K1, K2]{}, err
	}

	readTotal += read
	return readTotal, Join(key1, key2), nil
}

func (p pairKeyCodec[K1, K2]) Size(key Pair[K1, K2]) int {
	size := 0
	if key.key1 != nil {
		size += p.keyCodec1.SizeNonTerminal(*key.key1)
	}
	if key.key2 != nil {
		size += p.keyCodec2.Size(*key.key2)
	}
	return size
}

func (p pairKeyCodec[K1, K2]) Stringify(key Pair[K1, K2]) string {
	b := new(strings.Builder)
	b.WriteByte('(')
	if key.key1 != nil {
		b.WriteByte('"')
		b.WriteString(p.keyCodec1.Stringify(*key.key1))
		b.WriteByte('"')
	} else {
		b.WriteString("<nil>")
	}
	b.WriteString(", ")
	if key.key2 != nil {
		b.WriteByte('"')
		b.WriteString(p.keyCodec2.Stringify(*key.key2))
		b.WriteByte('"')
	} else {
		b.WriteString("<nil>")
	}
	b.WriteByte(')')
	return b.String()
}

func (p pairKeyCodec[K1, K2]) KeyType() string {
	return fmt.Sprintf("Pair[%s, %s]", p.keyCodec1.KeyType(), p.keyCodec2.KeyType())
}

func (p pairKeyCodec[K1, K2]) EncodeNonTerminal(buffer []byte, pair Pair[K1, K2]) (int, error) {
	writtenTotal := 0
	if pair.key1 != nil {
		written, err := p.keyCodec1.EncodeNonTerminal(buffer, *pair.key1)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if pair.key2 != nil {
		written, err := p.keyCodec2.EncodeNonTerminal(buffer[writtenTotal:], *pair.key2)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	return writtenTotal, nil
}

func (p pairKeyCodec[K1, K2]) DecodeNonTerminal(buffer []byte) (int, Pair[K1, K2], error) {
	readTotal := 0
	read, key1, err := p.keyCodec1.DecodeNonTerminal(buffer)
	if err != nil {
		return 0, Pair[K1, K2]{}, err
	}
	readTotal += read
	read, key2, err := p.keyCodec2.DecodeNonTerminal(buffer[read:])
	if err != nil {
		return 0, Pair[K1, K2]{}, err
	}

	readTotal += read
	return readTotal, Join(key1, key2), nil
}

func (p pairKeyCodec[K1, K2]) SizeNonTerminal(key Pair[K1, K2]) int {
	size := 0
	if key.key1 != nil {
		size += p.keyCodec1.SizeNonTerminal(*key.key1)
	}
	if key.key2 != nil {
		size += p.keyCodec2.SizeNonTerminal(*key.key2)
	}
	return size
}

// NewPairRange creates a new PairRange which will prefix over all the keys
// starting with the provided prefix.
func NewPairRange[K1, K2 any](prefix K1) *PairRange[K1, K2] {
	return &PairRange[K1, K2]{
		start: RangeBoundNone(PairPrefix[K1, K2](prefix)),
		end:   RangeBoundNextPrefixKey(PairPrefix[K1, K2](prefix)),
	}
}

// PairRange is an API that facilitates working with Pair iteration.
// It implements the Ranger API.
type PairRange[K1, K2 any] struct {
	start *RangeBound[Pair[K1, K2]]
	end   *RangeBound[Pair[K1, K2]]
	order Order

	err error
}

func (p *PairRange[K1, K2]) StartInclusive(k2 K2) *PairRange[K1, K2] {
	p.start = RangeBoundNone(Join(*p.start.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) StartExclusive(k2 K2) *PairRange[K1, K2] {
	p.start = RangeBoundNextKey(Join(*p.start.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) EndInclusive(k2 K2) *PairRange[K1, K2] {
	p.end = RangeBoundNextKey(Join(*p.end.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) EndExclusive(k2 K2) *PairRange[K1, K2] {
	p.end = RangeBoundNone(Join(*p.end.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) Descending() *PairRange[K1, K2] {
	p.order = OrderDescending
	return p
}

func (p *PairRange[K1, K2]) RangeValues() (start *RangeBound[Pair[K1, K2]], end *RangeBound[Pair[K1, K2]], order Order, err error) {
	if p.err != nil {
		return nil, nil, 0, err
	}
	return p.start, p.end, p.order, nil
}
