// Package statecheck verifies that every key and value in a module store
// decodes with the codecs of the module's collections schema.
//
// A store can hold entries written by an older version of a module whose
// encoding no longer matches the codec the current keeper reads with. Such an
// entry stays invisible until the keeper reads it, for example when an
// EndBlocker iterates a queue past an expiry and returns a decode error. The
// check here reads every entry right away, so a test can find the mismatch
// right after an upgrade.
package statecheck

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"cosmossdk.io/collections"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Module names a module store and the collections schema the module reads it with.
type Module struct {
	Name     string
	StoreKey storetypes.StoreKey
	Schema   collections.Schema
	// RawPrefixes lists the key prefixes under which the module keeps state
	// outside its collections. Check counts entries under these prefixes in the
	// result without decoding them. An entry outside every collection prefix
	// and every raw prefix still fails.
	RawPrefixes [][]byte
}

// Result reports how many entries of a module store Check matched to a
// collection and decoded, and how many it counted under a raw prefix of the
// module without decoding.
type Result struct {
	Name    string
	Checked int
	Raw     int
}

// Check decodes every entry of every module store and returns a result per
// module and an error that lists each entry that does not decode.
//
// An entry under the prefix of a collection fails when its key does not decode
// with the key codec of the collection or leaves bytes over, or when its value
// does not decode with the value codec of the collection. An entry outside
// every collection prefix fails unless it falls under a raw prefix of the
// module. A
// module whose schema lists no collections fails, since no entry of its store
// could be checked.
func Check(ctx sdk.Context, modules []Module) ([]Result, error) {
	results := make([]Result, 0, len(modules))
	var errs []error
	for _, m := range modules {
		res, modErrs := checkModule(ctx, m)
		results = append(results, res)
		errs = append(errs, modErrs...)
	}
	return results, errors.Join(errs...)
}

func checkModule(ctx sdk.Context, m Module) (Result, []error) {
	res := Result{Name: m.Name}
	colls := m.Schema.ListCollections()
	if len(colls) == 0 {
		return res, []error{fmt.Errorf("%s: schema lists no collections", m.Name)}
	}

	var errs []error
	it := ctx.KVStore(m.StoreKey).Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		key, value := it.Key(), it.Value()

		coll := collectionFor(colls, key)
		if coll == nil {
			if hasAnyPrefix(key, m.RawPrefixes) {
				res.Raw++
				continue
			}
			errs = append(errs, fmt.Errorf("%s: key %X is outside every collection prefix and every raw prefix", m.Name, key))
			continue
		}

		res.Checked++
		if err := decodeKey(coll, key[len(coll.GetPrefix()):]); err != nil {
			errs = append(errs, fmt.Errorf("%s/%s: key %X: %w", m.Name, coll.GetName(), key, err))
			continue
		}
		if _, err := coll.ValueCodec().Decode(value); err != nil {
			errs = append(errs, fmt.Errorf("%s/%s: key %X value %X: %w", m.Name, coll.GetName(), key, value, err))
		}
	}
	return res, errs
}

// hasAnyPrefix reports whether the key starts with one of the prefixes.
func hasAnyPrefix(key []byte, prefixes [][]byte) bool {
	for _, p := range prefixes {
		if bytes.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// collectionFor returns the collection whose prefix the key starts with, or
// nil. The schema builder rejects overlapping prefixes, so at most one matches.
func collectionFor(colls []collections.Collection, key []byte) collections.Collection {
	for _, c := range colls {
		if bytes.HasPrefix(key, c.GetPrefix()) {
			return c
		}
	}
	return nil
}

// decodeKey decodes a key, with its collection prefix removed, using the key
// codec of the collection, and requires the codec to consume every byte.
func decodeKey(coll collections.Collection, key []byte) error {
	kc := keyCodec(coll)
	out := kc.MethodByName("Decode").Call([]reflect.Value{reflect.ValueOf(key)})
	if errV := out[2]; !errV.IsNil() {
		return errV.Interface().(error)
	}
	if read := int(out[0].Int()); read != len(key) {
		return fmt.Errorf("%w: key codec %s read %d of %d bytes", collections.ErrEncoding, KeyType(coll), read, len(key))
	}
	return nil
}

// KeyType returns the KeyType of the key codec of a collection, for example
// "Pair[sdk.AccAddress,string]".
func KeyType(coll collections.Collection) string {
	return keyCodec(coll).MethodByName("KeyType").Call(nil)[0].String()
}

// CodecTypes returns the concrete Go types of the key codec and the value
// codec of a collection, for example "codec.AltValueCodec[cosmossdk.io/math.Int]".
// Unlike KeyType and ValueType, which a wrapping codec can share with the codec
// it wraps, the concrete type changes when a codec is swapped or a wrapper is
// added or removed.
func CodecTypes(coll collections.Collection) (key, value string) {
	m := innerMap(coll)
	key = fmt.Sprintf("%T", m.MethodByName("KeyCodec").Call(nil)[0].Interface())
	value = fmt.Sprintf("%T", m.MethodByName("ValueCodec").Call(nil)[0].Interface())
	return key, value
}

// keyCodec returns the key codec of a collection.
func keyCodec(coll collections.Collection) reflect.Value {
	return innerMap(coll).MethodByName("KeyCodec").Call(nil)[0]
}

// innerMap returns the collections.Map a schema collection wraps. The
// Collection interface does not expose the key codec or the typed value codec,
// but every collection in a schema wraps a collections.Map in an unexported
// field, and Map.KeyCodec and Map.ValueCodec are exported. The field is read
// through unsafe because reflect refuses to call methods on a value read from
// an unexported field.
func innerMap(coll collections.Collection) reflect.Value {
	v := reflect.ValueOf(coll)
	addressable := reflect.New(v.Type()).Elem()
	addressable.Set(v)
	field := addressable.Field(0)
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}
