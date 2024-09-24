package collections

import (
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/collections/codec"
	"cosmossdk.io/schema"
)

// Quad defines a multipart key composed of four keys.
type Quad[K1, K2, K3, K4 any] struct {
	k1 *K1
	k2 *K2
	k3 *K3
	k4 *K4
}

// Join4 instantiates a new Quad instance composed of the four provided keys, in order.
func Join4[K1, K2, K3, K4 any](k1 K1, k2 K2, k3 K3, k4 K4) Quad[K1, K2, K3, K4] {
	return Quad[K1, K2, K3, K4]{&k1, &k2, &k3, &k4}
}

// K1 returns the first part of the key. If nil, the zero value is returned.
func (t Quad[K1, K2, K3, K4]) K1() (x K1) {
	if t.k1 != nil {
		return *t.k1
	}
	return x
}

// K2 returns the second part of the key. If nil, the zero value is returned.
func (t Quad[K1, K2, K3, K4]) K2() (x K2) {
	if t.k2 != nil {
		return *t.k2
	}
	return x
}

// K3 returns the third part of the key. If nil, the zero value is returned.
func (t Quad[K1, K2, K3, K4]) K3() (x K3) {
	if t.k3 != nil {
		return *t.k3
	}
	return x
}

// K4 returns the fourth part of the key. If nil, the zero value is returned.
func (t Quad[K1, K2, K3, K4]) K4() (x K4) {
	if t.k4 != nil {
		return *t.k4
	}
	return x
}

// QuadPrefix creates a new Quad instance composed only of the first part of the key.
func QuadPrefix[K1, K2, K3, K4 any](k1 K1) Quad[K1, K2, K3, K4] {
	return Quad[K1, K2, K3, K4]{k1: &k1}
}

// QuadSuperPrefix creates a new Quad instance composed only of the first two parts of the key.
func QuadSuperPrefix[K1, K2, K3, K4 any](k1 K1, k2 K2) Quad[K1, K2, K3, K4] {
	return Quad[K1, K2, K3, K4]{k1: &k1, k2: &k2}
}

// QuadSuperPrefix3 creates a new Quad instance composed only of the first three parts of the key.
func QuadSuperPrefix3[K1, K2, K3, K4 any](k1 K1, k2 K2, k3 K3) Quad[K1, K2, K3, K4] {
	return Quad[K1, K2, K3, K4]{k1: &k1, k2: &k2, k3: &k3}
}

// QuadKeyCodec instantiates a new KeyCodec instance that can encode the Quad, given
// the KeyCodecs of the four parts of the key, in order.
func QuadKeyCodec[K1, K2, K3, K4 any](keyCodec1 codec.KeyCodec[K1], keyCodec2 codec.KeyCodec[K2], keyCodec3 codec.KeyCodec[K3], keyCodec4 codec.KeyCodec[K4]) codec.KeyCodec[Quad[K1, K2, K3, K4]] {
	return quadKeyCodec[K1, K2, K3, K4]{
		keyCodec1: keyCodec1,
		keyCodec2: keyCodec2,
		keyCodec3: keyCodec3,
		keyCodec4: keyCodec4,
	}
}

// NamedQuadKeyCodec instantiates a new KeyCodec instance that can encode the Quad, given
// the KeyCodecs of the four parts of the key, in order.
// The provided names are used to identify the parts of the key in the schema for indexing.
func NamedQuadKeyCodec[K1, K2, K3, K4 any](key1Name string, keyCodec1 codec.KeyCodec[K1], key2Name string, keyCodec2 codec.KeyCodec[K2], key3Name string, keyCodec3 codec.KeyCodec[K3], key4Name string, keyCodec4 codec.KeyCodec[K4]) codec.KeyCodec[Quad[K1, K2, K3, K4]] {
	return quadKeyCodec[K1, K2, K3, K4]{
		name1:     key1Name,
		keyCodec1: keyCodec1,
		name2:     key2Name,
		keyCodec2: keyCodec2,
		name3:     key3Name,
		keyCodec3: keyCodec3,
		name4:     key4Name,
		keyCodec4: keyCodec4,
	}
}

type quadKeyCodec[K1, K2, K3, K4 any] struct {
	name1, name2, name3, name4 string
	keyCodec1                  codec.KeyCodec[K1]
	keyCodec2                  codec.KeyCodec[K2]
	keyCodec3                  codec.KeyCodec[K3]
	keyCodec4                  codec.KeyCodec[K4]
}

type jsonQuadKey [4]json.RawMessage

func (t quadKeyCodec[K1, K2, K3, K4]) EncodeJSON(value Quad[K1, K2, K3, K4]) ([]byte, error) {
	json1, err := t.keyCodec1.EncodeJSON(*value.k1)
	if err != nil {
		return nil, err
	}

	json2, err := t.keyCodec2.EncodeJSON(*value.k2)
	if err != nil {
		return nil, err
	}

	json3, err := t.keyCodec3.EncodeJSON(*value.k3)
	if err != nil {
		return nil, err
	}

	json4, err := t.keyCodec4.EncodeJSON(*value.k4)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jsonQuadKey{json1, json2, json3, json4})
}

func (t quadKeyCodec[K1, K2, K3, K4]) DecodeJSON(b []byte) (Quad[K1, K2, K3, K4], error) {
	var jsonKey jsonQuadKey
	err := json.Unmarshal(b, &jsonKey)
	if err != nil {
		return Quad[K1, K2, K3, K4]{}, err
	}

	key1, err := t.keyCodec1.DecodeJSON(jsonKey[0])
	if err != nil {
		return Quad[K1, K2, K3, K4]{}, err
	}

	key2, err := t.keyCodec2.DecodeJSON(jsonKey[1])
	if err != nil {
		return Quad[K1, K2, K3, K4]{}, err
	}

	key3, err := t.keyCodec3.DecodeJSON(jsonKey[2])
	if err != nil {
		return Quad[K1, K2, K3, K4]{}, err
	}

	key4, err := t.keyCodec4.DecodeJSON(jsonKey[3])
	if err != nil {
		return Quad[K1, K2, K3, K4]{}, err
	}

	return Join4(key1, key2, key3, key4), nil
}

func (t quadKeyCodec[K1, K2, K3, K4]) Stringify(key Quad[K1, K2, K3, K4]) string {
	b := new(strings.Builder)
	b.WriteByte('(')
	if key.k1 != nil {
		b.WriteByte('"')
		b.WriteString(t.keyCodec1.Stringify(*key.k1))
		b.WriteByte('"')
	} else {
		b.WriteString("<nil>")
	}

	b.WriteString(", ")
	if key.k2 != nil {
		b.WriteByte('"')
		b.WriteString(t.keyCodec2.Stringify(*key.k2))
		b.WriteByte('"')
	} else {
		b.WriteString("<nil>")
	}

	b.WriteString(", ")
	if key.k3 != nil {
		b.WriteByte('"')
		b.WriteString(t.keyCodec3.Stringify(*key.k3))
		b.WriteByte('"')
	} else {
		b.WriteString("<nil>")
	}

	b.WriteString(", ")
	if key.k4 != nil {
		b.WriteByte('"')
		b.WriteString(t.keyCodec4.Stringify(*key.k4))
		b.WriteByte('"')
	} else {
		b.WriteString("<nil>")
	}

	b.WriteByte(')')
	return b.String()
}

func (t quadKeyCodec[K1, K2, K3, K4]) KeyType() string {
	return fmt.Sprintf("Quad[%s,%s,%s,%s]", t.keyCodec1.KeyType(), t.keyCodec2.KeyType(), t.keyCodec3.KeyType(), t.keyCodec4.KeyType())
}

func (t quadKeyCodec[K1, K2, K3, K4]) Encode(buffer []byte, key Quad[K1, K2, K3, K4]) (int, error) {
	writtenTotal := 0
	if key.k1 != nil {
		written, err := t.keyCodec1.EncodeNonTerminal(buffer, *key.k1)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if key.k2 != nil {
		written, err := t.keyCodec2.EncodeNonTerminal(buffer[writtenTotal:], *key.k2)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if key.k3 != nil {
		written, err := t.keyCodec3.EncodeNonTerminal(buffer[writtenTotal:], *key.k3)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if key.k4 != nil {
		written, err := t.keyCodec4.Encode(buffer[writtenTotal:], *key.k4)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	return writtenTotal, nil
}

func (t quadKeyCodec[K1, K2, K3, K4]) Decode(buffer []byte) (int, Quad[K1, K2, K3, K4], error) {
	readTotal := 0
	read, key1, err := t.keyCodec1.DecodeNonTerminal(buffer)
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	read, key2, err := t.keyCodec2.DecodeNonTerminal(buffer[readTotal:])
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	read, key3, err := t.keyCodec3.DecodeNonTerminal(buffer[readTotal:])
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	read, key4, err := t.keyCodec4.Decode(buffer[readTotal:])
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	return readTotal, Join4(key1, key2, key3, key4), nil
}

func (t quadKeyCodec[K1, K2, K3, K4]) Size(key Quad[K1, K2, K3, K4]) int {
	size := 0
	if key.k1 != nil {
		size += t.keyCodec1.SizeNonTerminal(*key.k1)
	}
	if key.k2 != nil {
		size += t.keyCodec2.SizeNonTerminal(*key.k2)
	}
	if key.k3 != nil {
		size += t.keyCodec3.SizeNonTerminal(*key.k3)
	}
	if key.k4 != nil {
		size += t.keyCodec4.Size(*key.k4)
	}
	return size
}

func (t quadKeyCodec[K1, K2, K3, K4]) EncodeNonTerminal(buffer []byte, key Quad[K1, K2, K3, K4]) (int, error) {
	writtenTotal := 0
	if key.k1 != nil {
		written, err := t.keyCodec1.EncodeNonTerminal(buffer, *key.k1)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if key.k2 != nil {
		written, err := t.keyCodec2.EncodeNonTerminal(buffer[writtenTotal:], *key.k2)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if key.k3 != nil {
		written, err := t.keyCodec3.EncodeNonTerminal(buffer[writtenTotal:], *key.k3)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	if key.k4 != nil {
		written, err := t.keyCodec4.EncodeNonTerminal(buffer[writtenTotal:], *key.k4)
		if err != nil {
			return 0, err
		}
		writtenTotal += written
	}
	return writtenTotal, nil
}

func (t quadKeyCodec[K1, K2, K3, K4]) DecodeNonTerminal(buffer []byte) (int, Quad[K1, K2, K3, K4], error) {
	readTotal := 0
	read, key1, err := t.keyCodec1.DecodeNonTerminal(buffer)
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	read, key2, err := t.keyCodec2.DecodeNonTerminal(buffer[readTotal:])
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	read, key3, err := t.keyCodec3.DecodeNonTerminal(buffer[readTotal:])
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	read, key4, err := t.keyCodec4.DecodeNonTerminal(buffer[readTotal:])
	if err != nil {
		return 0, Quad[K1, K2, K3, K4]{}, err
	}
	readTotal += read
	return readTotal, Join4(key1, key2, key3, key4), nil
}

func (t quadKeyCodec[K1, K2, K3, K4]) SizeNonTerminal(key Quad[K1, K2, K3, K4]) int {
	size := 0
	if key.k1 != nil {
		size += t.keyCodec1.SizeNonTerminal(*key.k1)
	}
	if key.k2 != nil {
		size += t.keyCodec2.SizeNonTerminal(*key.k2)
	}
	if key.k3 != nil {
		size += t.keyCodec3.SizeNonTerminal(*key.k3)
	}
	if key.k4 != nil {
		size += t.keyCodec4.SizeNonTerminal(*key.k4)
	}
	return size
}

func (t quadKeyCodec[K1, K2, K3, K4]) SchemaCodec() (codec.SchemaCodec[Quad[K1, K2, K3, K4]], error) {
	field1, err := getNamedKeyField(t.keyCodec1, t.name1)
	if err != nil {
		return codec.SchemaCodec[Quad[K1, K2, K3, K4]]{}, fmt.Errorf("error getting key1 field: %w", err)
	}

	field2, err := getNamedKeyField(t.keyCodec2, t.name2)
	if err != nil {
		return codec.SchemaCodec[Quad[K1, K2, K3, K4]]{}, fmt.Errorf("error getting key2 field: %w", err)
	}

	field3, err := getNamedKeyField(t.keyCodec3, t.name3)
	if err != nil {
		return codec.SchemaCodec[Quad[K1, K2, K3, K4]]{}, fmt.Errorf("error getting key3 field: %w", err)
	}

	field4, err := getNamedKeyField(t.keyCodec4, t.name4)
	if err != nil {
		return codec.SchemaCodec[Quad[K1, K2, K3, K4]]{}, fmt.Errorf("error getting key4 field: %w", err)
	}

	return codec.SchemaCodec[Quad[K1, K2, K3, K4]]{
		Fields: []schema.Field{field1, field2, field3, field4},
	}, nil
}

// NewPrefixUntilQuadRange defines a collection query which ranges until the provided Quad prefix.
// Unstable: this API might change in the future.
func NewPrefixUntilQuadRange[K1, K2, K3, K4 any](k1 K1) Ranger[Quad[K1, K2, K3, K4]] {
	key := QuadPrefix[K1, K2, K3, K4](k1)
	return &Range[Quad[K1, K2, K3, K4]]{
		end: RangeKeyPrefixEnd(key),
	}
}

// NewPrefixedQuadRange provides a Range for all keys prefixed with the given
// first part of the Quad key.
func NewPrefixedQuadRange[K1, K2, K3, K4 any](k1 K1) Ranger[Quad[K1, K2, K3, K4]] {
	key := QuadPrefix[K1, K2, K3, K4](k1)
	return &Range[Quad[K1, K2, K3, K4]]{
		start: RangeKeyExact(key),
		end:   RangeKeyPrefixEnd(key),
	}
}

// NewSuperPrefixedQuadRange provides a Range for all keys prefixed with the given
// first and second parts of the Quad key.
func NewSuperPrefixedQuadRange[K1, K2, K3, K4 any](k1 K1, k2 K2) Ranger[Quad[K1, K2, K3, K4]] {
	key := QuadSuperPrefix[K1, K2, K3, K4](k1, k2)
	return &Range[Quad[K1, K2, K3, K4]]{
		start: RangeKeyExact(key),
		end:   RangeKeyPrefixEnd(key),
	}
}

// NewSuperPrefixedQuadRange3 provides a Range for all keys prefixed with the given
// first, second and third parts of the Quad key.
func NewSuperPrefixedQuadRange3[K1, K2, K3, K4 any](k1 K1, k2 K2, k3 K3) Ranger[Quad[K1, K2, K3, K4]] {
	key := QuadSuperPrefix3[K1, K2, K3, K4](k1, k2, k3)
	return &Range[Quad[K1, K2, K3, K4]]{
		start: RangeKeyExact(key),
		end:   RangeKeyPrefixEnd(key),
	}
}
