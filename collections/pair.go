package collections

import "fmt"

type Pair[K1, K2 any] struct {
	key1 *K1
	key2 *K2
}

func Join[K1, K2 any](key1 K1, key2 K2) Pair[K1, K2] {
	return Pair[K1, K2]{
		key1: &key1,
		key2: &key2,
	}
}

func PairKeyCodec[K1, K2 any](keyCodec1 KeyCodec[K1], keyCodec2 KeyCodec[K2]) KeyCodec[Pair[K1, K2]] {
	return pairKeyCodec[K1, K2]{}
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
	read, key1, err := p.keyCodec1.Decode(buffer)
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
		size += p.keyCodec1.Size(*key.key1)
	}
	if key.key2 != nil {
		size += p.keyCodec2.Size(*key.key2)
	}
	return size
}

func (p pairKeyCodec[K1, K2]) Stringify(key Pair[K1, K2]) string {
	panic("impl")
}

func (p pairKeyCodec[K1, K2]) KeyType() string {
	return fmt.Sprintf("Pair[%s, %s]", p.keyCodec1.KeyType(), p.keyCodec2.KeyType())
}

func (p pairKeyCodec[K1, K2]) EncodeNonTerminal(buffer []byte, key Pair[K1, K2]) (int, error) {
	panic("impl")
}

func (p pairKeyCodec[K1, K2]) DecodeNonTerminal(buffer []byte) (int, Pair[K1, K2], error) {
	panic("impl")
}

func (p pairKeyCodec[K1, K2]) SizeNonTerminal(key Pair[K1, K2]) int {
	panic("impl")
}

type PairRange[K1, K2 any] struct {
	prefix *K1
	start  *Bound[K2]
	end    *Bound[K2]
	order  Order
}

func (p *PairRange[K1, K2]) Prefix(k1 K1) *PairRange[K1, K2] {
	p.prefix = &k1
	return p
}

func (p *PairRange[K1, K2]) StartInclusive(k2 K2) *PairRange[K1, K2] {
	p.start = BoundInclusive(k2)
	return p
}

func (p *PairRange[K1, K2]) StartExclusive(k2 K2) *PairRange[K1, K2] {
	p.start = BoundExclusive(k2)
	return p
}

func (p *PairRange[K1, K2]) EndInclusive(k2 K2) *PairRange[K1, K2] {
	p.end = BoundInclusive(k2)
	return p
}

func (p *PairRange[K1, K2]) EndExclusive(k2 K2) *PairRange[K1, K2] {
	p.end = BoundExclusive(k2)
	return p
}

func (p *PairRange[K1, K2]) Descending() *PairRange[K1, K2] {
	p.order = OrderDescending
	return p
}

func (p *PairRange[K1, K2]) RangeValues() (prefix *Pair[K1, K2], start *Bound[Pair[K1, K2]], end *Bound[Pair[K1, K2]], order Order, err error) {
	order = p.order
	panic("impl")
}
