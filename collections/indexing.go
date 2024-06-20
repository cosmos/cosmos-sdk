package collections

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/btree"

	"cosmossdk.io/collections/codec"
	indexerbase "cosmossdk.io/indexer/base"
)

type IndexingOptions struct {
	RetainDeletionsFor []string
}

func (s Schema) ModuleDecoder(opts IndexingOptions) (indexerbase.ModuleDecoder, error) {
	decoder := moduleDecoder{
		collectionLookup: &btree.Map[string, *collDecoder]{},
	}

	retainDeletions := make(map[string]bool)
	for _, collName := range opts.RetainDeletionsFor {
		retainDeletions[collName] = true
	}

	var objectTypes []indexerbase.ObjectType
	for _, collName := range s.collectionsOrdered {
		coll := s.collectionsByName[collName]

		// skip secondary indexes
		if coll.isSecondaryIndex() {
			continue
		}

		ld := coll.logicalDecoder()

		if !retainDeletions[coll.GetName()] {
			ld.objectType.RetainDeletions = true
		}

		objectTypes = append(objectTypes, ld.objectType)
		decoder.collectionLookup.Set(string(coll.GetPrefix()), &collDecoder{
			Collection:     coll,
			logicalDecoder: ld,
		})
	}

	return indexerbase.ModuleDecoder{
		Schema: indexerbase.ModuleSchema{
			ObjectTypes: objectTypes,
		},
		KVDecoder: decoder.decodeKV,
	}, nil
}

type moduleDecoder struct {
	collectionLookup *btree.Map[string, *collDecoder]
}

func (m moduleDecoder) decodeKV(key, value []byte) (indexerbase.ObjectUpdate, bool, error) {
	ks := string(key)
	var cd *collDecoder
	m.collectionLookup.Descend(ks, func(prefix string, cur *collDecoder) bool {
		bytesPrefix := cur.GetPrefix()
		if bytes.HasPrefix(key, bytesPrefix) {
			cd = cur
			return true
		}
		return false
	})
	if cd == nil {
		return indexerbase.ObjectUpdate{}, false, nil
	}

	return cd.decodeKVPair(key, value, false)
}

type collDecoder struct {
	Collection
	logicalDecoder
}

func (c collDecoder) decodeKVPair(key, value []byte, delete bool) (indexerbase.ObjectUpdate, bool, error) {
	// strip prefix
	key = key[len(c.GetPrefix()):]

	k, err := c.keyDecoder(key)
	if err != nil {
		return indexerbase.ObjectUpdate{
			TypeName: c.GetName(),
		}, true, err

	}

	if delete {
		return indexerbase.ObjectUpdate{
			TypeName: c.GetName(),
			Key:      k,
			Delete:   true,
		}, true, nil
	}

	v, err := c.valueDecoder(key)
	if err != nil {
		return indexerbase.ObjectUpdate{
			TypeName: c.GetName(),
			Key:      k,
		}, true, err

	}

	return indexerbase.ObjectUpdate{
		TypeName: c.GetName(),
		Key:      k,
		Value:    v,
	}, true, nil
}

func (c collectionImpl[K, V]) logicalDecoder() logicalDecoder {
	res := logicalDecoder{}

	if indexable, ok := c.m.kc.(codec.IndexableCodec); ok {
		keyDecoder := indexable.LogicalDecoder()
		res.objectType.KeyFields = keyDecoder.Fields
		res.keyDecoder = keyDecoder.Decode
	} else {
		fields, decoder := fallbackDecoder[K](func(bz []byte) (any, error) {
			_, k, err := c.m.kc.Decode(bz)
			return k, err
		})
		res.objectType.KeyFields = fields
		res.keyDecoder = decoder
	}
	ensureFieldNames(c.m.kc, "key", res.objectType.KeyFields)

	if indexable, ok := c.m.vc.(codec.IndexableCodec); ok {
		valueDecoder := indexable.LogicalDecoder()
		res.objectType.KeyFields = valueDecoder.Fields
		res.valueDecoder = valueDecoder.Decode
	} else {
		fields, decoder := fallbackDecoder[V](func(bz []byte) (any, error) {
			v, err := c.m.vc.Decode(bz)
			return v, err
		})
		res.objectType.ValueFields = fields
		res.valueDecoder = decoder
	}
	ensureFieldNames(c.m.vc, "value", res.objectType.ValueFields)

	return res
}

func fallbackDecoder[T any](decode func([]byte) (any, error)) ([]indexerbase.Field, func([]byte) (any, error)) {
	var t T
	kind := indexerbase.KindForGoValue(t)
	if err := kind.Validate(); err == nil {
		return []indexerbase.Field{{Kind: kind}}, decode
	} else {
		return []indexerbase.Field{{Kind: indexerbase.JSONKind}}, func(b []byte) (any, error) {
			t, err := decode(b)
			bz, err := json.Marshal(t)
			return json.RawMessage(bz), err
		}
	}
}

func ensureFieldNames(x any, defaultName string, cols []indexerbase.Field) {
	var names []string = nil
	if hasName, ok := x.(interface{ Name() string }); ok {
		name := hasName.Name()
		if name != "" {
			names = strings.Split(hasName.Name(), ",")
		}
	}
	for i, col := range cols {
		if names != nil && i < len(names) {
			col.Name = names[i]
		} else {
			if col.Name == "" {
				if i == 0 && len(cols) == 1 {
					col.Name = defaultName
				} else {
					col.Name = fmt.Sprintf("%s%d", defaultName, i+1)
				}
			}
		}
		cols[i] = col
	}
}
