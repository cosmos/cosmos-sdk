package collections

import (
	"context"
	"encoding/json"
	"fmt"
	io "io"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Map represents the basic collections object.
// It is used to map arbitrary keys to arbitrary
// objects.
type Map[K, V any] struct {
	kc KeyCodec[K]
	vc ValueCodec[V]

	sk     storetypes.StoreKey
	prefix []byte
	name   string
}

// NewMap returns a Map given a StoreKey, a Prefix, human-readable name and the relative value and key encoders.
// Name and prefix must be unique within the schema and name must match the format specified by NameRegex, or
// else this method will panic.
func NewMap[K, V any](
	schema Schema,
	prefix Prefix,
	name string,
	keyCodec KeyCodec[K],
	valueCodec ValueCodec[V],
) Map[K, V] {
	m := newMap(schema, prefix, name, keyCodec, valueCodec)
	schema.addCollection(m)
	return m
}

func newMap[K, V any](
	schema Schema, prefix Prefix, name string,
	keyCodec KeyCodec[K], valueCodec ValueCodec[V],
) Map[K, V] {
	return Map[K, V]{
		kc:     keyCodec,
		vc:     valueCodec,
		sk:     schema.storeKey,
		prefix: prefix.Bytes(),
		name:   name,
	}
}

func (m Map[K, V]) getName() string {
	return m.name
}

func (m Map[K, V]) getPrefix() []byte {
	return m.prefix
}

// Set maps the provided value to the provided key in the store.
// Errors with ErrEncoding if key or value encoding fails.
func (m Map[K, V]) Set(ctx context.Context, key K, value V) error {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)

	if err != nil {
		return err
	}

	valueBytes, err := m.vc.Encode(value)
	if err != nil {
		return fmt.Errorf("%w: value encode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}

	store, err := m.getStore(ctx)
	if err != nil {
		return err
	}
	store.Set(bytesKey, valueBytes)
	return nil
}

// Get returns the value associated with the provided key,
// errors with ErrNotFound if the key does not exist, or
// with ErrEncoding if the key or value decoding fails.
func (m Map[K, V]) Get(ctx context.Context, key K) (V, error) {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		var v V
		return v, err
	}

	store, err := m.getStore(ctx)
	if err != nil {
		var v V
		return v, err
	}
	valueBytes := store.Get(bytesKey)
	if valueBytes == nil {
		var v V
		return v, fmt.Errorf("%w: key '%s' of type %s", ErrNotFound, m.kc.Stringify(key), m.vc.ValueType())
	}

	v, err := m.vc.Decode(valueBytes)
	if err != nil {
		return v, fmt.Errorf("%w: value decode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}
	return v, nil
}

// Has reports whether the key is present in storage or not.
// Errors with ErrEncoding if key encoding fails.
func (m Map[K, V]) Has(ctx context.Context, key K) (bool, error) {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return false, err
	}
	store, err := m.getStore(ctx)
	if err != nil {
		return false, err
	}
	return store.Has(bytesKey), nil
}

// Remove removes the key from the storage.
// Errors with ErrEncoding if key encoding fails.
// If the key does not exist then this is a no-op.
func (m Map[K, V]) Remove(ctx context.Context, key K) error {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return err
	}
	store, err := m.getStore(ctx)
	if err != nil {
		return err
	}
	store.Delete(bytesKey)
	return nil
}

// Iterate provides an Iterator over K and V. It accepts a Ranger interface.
// A nil ranger equals to iterate over all the keys in ascending order.
func (m Map[K, V]) Iterate(ctx context.Context, ranger Ranger[K]) (Iterator[K, V], error) {
	return iteratorFromRanger(ctx, m, ranger)
}

func (m Map[K, V]) getStore(ctx context.Context) (storetypes.KVStore, error) {
	provider, ok := ctx.(StorageProvider)
	if !ok {
		return nil, fmt.Errorf("context is not a StorageProvider: underlying type %T", ctx)
	}
	return provider.KVStore(m.sk), nil
}

func (m Map[K, V]) defaultGenesis(writer io.Writer) error {
	_, err := writer.Write([]byte(`[]`))
	return err
}

func (m Map[K, V]) validateGenesis(reader io.Reader) error {
	return m.doDecodeJson(reader, func(key K, value V) error {
		return nil
	})
}

type jsonMapEntry struct {
	Key   json.RawMessage `json:"key"`
	Value json.RawMessage `json:"value"`
}

func (m Map[K, V]) importGenesis(ctx context.Context, reader io.Reader) error {
	return m.doDecodeJson(reader, func(key K, value V) error {
		return m.Set(ctx, key, value)
	})
}

func (m Map[K, V]) exportGenesis(ctx context.Context, writer io.Writer) error {
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	it, err := m.Iterate(ctx, nil)
	if err != nil {
		return err
	}

	start := true
	for ; it.Valid(); it.Next() {
		if !start {
			_, err = writer.Write([]byte(",\n"))
			if err != nil {
				return err
			}
		}
		start = false

		key, err := it.Key()
		if err != nil {
			return err
		}

		keyBz, err := m.kc.EncodeJSON(key)
		if err != nil {
			return err
		}

		value, err := it.Value()
		if err != nil {
			return err
		}

		valueBz, err := m.vc.EncodeJSON(value)
		if err != nil {
			return err
		}

		entry := jsonMapEntry{
			Key:   keyBz,
			Value: valueBz,
		}

		bz, err := json.Marshal(entry)
		if err != nil {

		}

		_, err = writer.Write(bz)
		if err != nil {
			return err
		}
	}

	_, err = writer.Write([]byte("]"))
	return err
}

func (m Map[K, V]) doDecodeJson(reader io.Reader, onEntry func(key K, value V) error) error {
	decoder := json.NewDecoder(reader)
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	if token != json.Delim('[') {
		return fmt.Errorf("expected [ got %s", token)
	}

	for decoder.More() {
		var rawJson json.RawMessage
		err := decoder.Decode(&rawJson)
		if err != nil {
			return err
		}

		var mapEntry jsonMapEntry
		err = json.Unmarshal(rawJson, &mapEntry)
		if err != nil {
			return err
		}

		key, err := m.kc.DecodeJSON(mapEntry.Key)
		if err != nil {
			return err
		}

		value, err := m.vc.DecodeJSON(mapEntry.Value)
		if err != nil {
			return err
		}

		err = onEntry(key, value)
		if err != nil {
			return err
		}
	}

	token, err = decoder.Token()
	if err != nil {
		return err
	}

	if token != json.Delim(']') {
		return fmt.Errorf("expected ] got %s", token)
	}

	return nil
}

func encodeKeyWithPrefix[K any](prefix []byte, kc KeyCodec[K], key K) ([]byte, error) {
	prefixLen := len(prefix)
	// preallocate buffer
	keyBytes := make([]byte, prefixLen+kc.Size(key))
	// put prefix
	copy(keyBytes, prefix)
	// put key
	_, err := kc.Encode(keyBytes[prefixLen:], key)
	if err != nil {
		return nil, fmt.Errorf("%w: key encode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}
	return keyBytes, nil
}
