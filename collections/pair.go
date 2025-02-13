package collections

import (
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/collections/codec"
	"cosmossdk.io/schema"
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

// PairKeyCodec instantiates a new KeyCodec instance that can encode the Pair, given the KeyCodec of the
// first part of the key and the KeyCodec of the second part of the key.
func PairKeyCodec[K1, K2 any](keyCodec1 codec.KeyCodec[K1], keyCodec2 codec.KeyCodec[K2]) codec.KeyCodec[Pair[K1, K2]] {
	return pairKeyCodec[K1, K2]{
		keyCodec1: keyCodec1,
		keyCodec2: keyCodec2,
	}
}

// NamedPairKeyCodec instantiates a new KeyCodec instance that can encode the Pair, given the KeyCodec of the
// first part of the key and the KeyCodec of the second part of the key.
// It also provides names for the keys which are used for indexing purposes.
func NamedPairKeyCodec[K1, K2 any](key1Name string, keyCodec1 codec.KeyCodec[K1], key2Name string, keyCodec2 codec.KeyCodec[K2]) codec.KeyCodec[Pair[K1, K2]] {
	return pairKeyCodec[K1, K2]{
		key1Name:  key1Name,
		key2Name:  key2Name,
		keyCodec1: keyCodec1,
		keyCodec2: keyCodec2,
	}
}

type pairKeyCodec[K1, K2 any] struct {
	key1Name, key2Name string
	keyCodec1          codec.KeyCodec[K1]
	keyCodec2          codec.KeyCodec[K2]
}

func (p pairKeyCodec[K1, K2]) KeyCodec1() codec.KeyCodec[K1] { return p.keyCodec1 }

func (p pairKeyCodec[K1, K2]) KeyCodec2() codec.KeyCodec[K2] { return p.keyCodec2 }

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

// GENESIS

type jsonPairKey [2]json.RawMessage

func (p pairKeyCodec[K1, K2]) EncodeJSON(v Pair[K1, K2]) ([]byte, error) {
	k1Json, err := p.keyCodec1.EncodeJSON(v.K1())
	if err != nil {
		return nil, err
	}
	k2Json, err := p.keyCodec2.EncodeJSON(v.K2())
	if err != nil {
		return nil, err
	}
	return json.Marshal(jsonPairKey{k1Json, k2Json})
}

func (p pairKeyCodec[K1, K2]) DecodeJSON(b []byte) (Pair[K1, K2], error) {
	pairJSON := jsonPairKey{}
	err := json.Unmarshal(b, &pairJSON)
	if err != nil {
		return Pair[K1, K2]{}, err
	}

	k1, err := p.keyCodec1.DecodeJSON(pairJSON[0])
	if err != nil {
		return Pair[K1, K2]{}, err
	}
	k2, err := p.keyCodec2.DecodeJSON(pairJSON[1])
	if err != nil {
		return Pair[K1, K2]{}, err
	}

	return Join(k1, k2), nil
}

func (p pairKeyCodec[K1, K2]) Name() string {
	return fmt.Sprintf("%s,%s", p.key1Name, p.key2Name)
}

func (p pairKeyCodec[K1, K2]) SchemaCodec() (codec.SchemaCodec[Pair[K1, K2]], error) {
	field1, err := getNamedKeyField(p.keyCodec1, p.key1Name)
	if err != nil {
		return codec.SchemaCodec[Pair[K1, K2]]{}, fmt.Errorf("error getting key1 field: %w", err)
	}

	field2, err := getNamedKeyField(p.keyCodec2, p.key2Name)
	if err != nil {
		return codec.SchemaCodec[Pair[K1, K2]]{}, fmt.Errorf("error getting key2 field: %w", err)
	}

	codec1, err := codec.KeySchemaCodec(p.keyCodec1)
	if err != nil {
		return codec.SchemaCodec[Pair[K1, K2]]{}, fmt.Errorf("error getting key1 schema codec: %w", err)
	}

	codec2, err := codec.KeySchemaCodec(p.keyCodec2)
	if err != nil {
		return codec.SchemaCodec[Pair[K1, K2]]{}, fmt.Errorf("error getting key2 schema codec: %w", err)
	}

	return codec.SchemaCodec[Pair[K1, K2]]{
		Fields: []schema.Field{field1, field2},
		ToSchemaType: func(pair Pair[K1, K2]) (any, error) {
			k1, err := toKeySchemaType(codec1, pair.K1())
			if err != nil {
				return nil, err
			}
			k2, err := toKeySchemaType(codec2, pair.K2())
			if err != nil {
				return nil, err
			}
			return []interface{}{k1, k2}, nil
		},
		FromSchemaType: func(a any) (Pair[K1, K2], error) {
			aSlice, ok := a.([]interface{})
			if !ok || len(aSlice) != 2 {
				return Pair[K1, K2]{}, fmt.Errorf("expected slice of length 2, got %T", a)
			}
			k1, err := fromKeySchemaType(codec1, aSlice[0])
			if err != nil {
				return Pair[K1, K2]{}, err
			}
			k2, err := fromKeySchemaType(codec2, aSlice[1])
			if err != nil {
				return Pair[K1, K2]{}, err
			}
			return Join(k1, k2), nil
		},
	}, nil
}

func getNamedKeyField[T any](keyCdc codec.KeyCodec[T], name string) (schema.Field, error) {
	keySchema, err := codec.KeySchemaCodec(keyCdc)
	if err != nil {
		return schema.Field{}, err
	}
	if len(keySchema.Fields) != 1 {
		return schema.Field{}, fmt.Errorf("key schema in composite key has more than one field, got %v", keySchema.Fields)
	}
	field := keySchema.Fields[0]
	field.Name = name
	return field, nil
}

func toKeySchemaType[T any](cdc codec.SchemaCodec[T], key T) (any, error) {
	if cdc.ToSchemaType != nil {
		return cdc.ToSchemaType(key)
	}
	return key, nil
}

func fromKeySchemaType[T any](cdc codec.SchemaCodec[T], key any) (T, error) {
	if cdc.FromSchemaType != nil {
		return cdc.FromSchemaType(key)
	}
	tKey, ok := key.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("expected type %T, got %T", zero, key)
	}
	return tKey, nil
}

// NewPrefixUntilPairRange defines a collection query which ranges until the provided Pair prefix.
// Unstable: this API might change in the future.
func NewPrefixUntilPairRange[K1, K2 any](prefix K1) *PairRange[K1, K2] {
	return &PairRange[K1, K2]{end: RangeKeyPrefixEnd(PairPrefix[K1, K2](prefix))}
}

// NewPrefixedPairRange creates a new PairRange which will prefix over all the keys
// starting with the provided prefix.
func NewPrefixedPairRange[K1, K2 any](prefix K1) *PairRange[K1, K2] {
	return &PairRange[K1, K2]{
		start: RangeKeyExact(PairPrefix[K1, K2](prefix)),
		end:   RangeKeyPrefixEnd(PairPrefix[K1, K2](prefix)),
	}
}

// PairRange is an API that facilitates working with Pair iteration.
// It implements the Ranger API.
// Unstable: API and methods are currently unstable.
type PairRange[K1, K2 any] struct {
	start *RangeKey[Pair[K1, K2]]
	end   *RangeKey[Pair[K1, K2]]
	order Order

	err error
}

func (p *PairRange[K1, K2]) StartInclusive(k2 K2) *PairRange[K1, K2] {
	p.start = RangeKeyExact(Join(*p.start.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) StartExclusive(k2 K2) *PairRange[K1, K2] {
	p.start = RangeKeyNext(Join(*p.start.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) EndInclusive(k2 K2) *PairRange[K1, K2] {
	p.end = RangeKeyNext(Join(*p.end.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) EndExclusive(k2 K2) *PairRange[K1, K2] {
	p.end = RangeKeyExact(Join(*p.end.key.key1, k2))
	return p
}

func (p *PairRange[K1, K2]) Descending() *PairRange[K1, K2] {
	p.order = OrderDescending
	return p
}

func (p *PairRange[K1, K2]) RangeValues() (start, end *RangeKey[Pair[K1, K2]], order Order, err error) {
	if p.err != nil {
		return nil, nil, 0, err
	}
	return p.start, p.end, p.order, nil
}
