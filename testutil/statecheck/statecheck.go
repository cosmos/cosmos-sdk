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
}

// Check decodes every entry of every module store and returns an error that
// lists each entry that does not decode. An entry under the prefix of a
// collection fails when its key does not decode with the key codec of the
// collection or leaves bytes over, or when its value does not decode with the
// value codec of the collection. Entries outside every collection prefix are
// not checked.
func Check(ctx sdk.Context, modules []Module) error {
	var errs []error
	for _, m := range modules {
		errs = append(errs, checkModule(ctx, m)...)
	}
	return errors.Join(errs...)
}

func checkModule(ctx sdk.Context, m Module) []error {
	colls := m.Schema.ListCollections()

	var errs []error
	it := ctx.KVStore(m.StoreKey).Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		key, value := it.Key(), it.Value()

		coll := collectionFor(colls, key)
		if coll == nil {
			// Some modules keep part of their state outside collections,
			// with no codec to check it against.
			continue
		}

		if err := decodeKey(coll, key[len(coll.GetPrefix()):]); err != nil {
			errs = append(errs, fmt.Errorf("%s/%s: key %X: %w", m.Name, coll.GetName(), key, err))
			continue
		}
		if _, err := coll.ValueCodec().Decode(value); err != nil {
			errs = append(errs, fmt.Errorf("%s/%s: key %X value %X: %w", m.Name, coll.GetName(), key, value, err))
		}
	}
	return errs
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

// keyCodec returns the key codec of a collection. The Collection interface does
// not expose the key codec, but every collection in a schema wraps a
// collections.Map in an unexported field, and Map.KeyCodec is exported. The
// field is read through unsafe because reflect refuses to call methods on a
// value read from an unexported field.
func keyCodec(coll collections.Collection) reflect.Value {
	v := reflect.ValueOf(coll)
	addressable := reflect.New(v.Type()).Elem()
	addressable.Set(v)
	field := addressable.Field(0)
	m := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	return m.MethodByName("KeyCodec").Call(nil)[0]
}
