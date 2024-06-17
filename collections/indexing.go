package collections

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tidwall/btree"

	indexerbase "cosmossdk.io/indexer/base"
)

type IndexingOptions struct {
	RetainDeletionsFor []string
}

func (s Schema) ModuleDecoder(opts IndexingOptions) (indexerbase.ModuleDecoder, error) {
	decoder := moduleDecoder{
		lookup: &btree.Map[string, *collDecoder]{},
	}

	var objectTypes []indexerbase.ObjectType
	for _, collName := range s.collectionsOrdered {
		coll := s.collectionsByName[collName]
		if coll.isIndex() {
			continue
		}

		schema := coll.getTableSchema()
		objectTypes = append(objectTypes, schema)
		decoder.lookup.Set(string(coll.GetPrefix()), &collDecoder{Collection: coll})
	}
	return indexerbase.ModuleDecoder{
		Schema: indexerbase.ModuleSchema{
			ObjectTypes: objectTypes,
		},
		KVDecoder: decoder.decodeKV,
	}, nil
}

type moduleDecoder struct {
	lookup *btree.Map[string, *collDecoder]
}

func (m moduleDecoder) decodeKV(key, value []byte) (indexerbase.ObjectUpdate, bool, error) {
	ks := string(key)
	var cd *collDecoder
	m.lookup.Descend(ks, func(prefix string, cur *collDecoder) bool {
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
}

//type moduleStateDecoder struct {
//	schema           Schema
//	collectionsIndex btree.BTree
//}
//
//func (m moduleStateDecoder) getCollectionForKey(key []byte) Collection {
//	panic("implement me")
//}
//
//func (m moduleStateDecoder) DecodeSet(key, value []byte) (indexerbase.EntityUpdate, bool, error) {
//	coll := m.getCollectionForKey(key)
//	if coll == nil {
//		return indexerbase.EntityUpdate{}, false, nil
//	}
//
//	return coll.decodeKVPair(key, value)
//}
//
//func (m moduleStateDecoder) DecodeDelete(key []byte) (indexerbase.EntityDelete, bool, error) {
//	coll := m.getCollectionForKey(key)
//	return coll.decodeDelete(key)
//}

func (c collectionImpl[K, V]) getTableSchema() indexerbase.ObjectType {
	var keyFields []indexerbase.Field
	var valueFields []indexerbase.Field

	if hasSchema, ok := c.m.kc.(IndexableCodec); ok {
		keyFields = hasSchema.SchemaFields()
	} else {
		var k K
		keyFields, _ = extractFields(k)
	}
	ensureNames(c.m.kc, "key", keyFields)

	if hasSchema, ok := c.m.vc.(IndexableCodec); ok {
		valueFields = hasSchema.SchemaFields()
	} else {
		var v V
		valueFields, _ = extractFields(v)
	}
	ensureNames(c.m.vc, "value", valueFields)

	return indexerbase.ObjectType{
		Name:        c.GetName(),
		KeyFields:   keyFields,
		ValueFields: valueFields,
	}
}

func extractFields(x any) ([]indexerbase.Field, func(any) any) {
	if hasSchema, ok := x.(IndexableCodec); ok {
		return hasSchema.SchemaFields(), nil
	}

	ty := indexerbase.KindForGoValue(x)
	if ty > 0 {
		return []indexerbase.Field{{Kind: ty}}, nil
	}

	panic(fmt.Errorf("unsupported type %T", x))
}

func ensureNames(x any, defaultName string, cols []indexerbase.Field) {
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

func (c collectionImpl[K, V]) decodeKVPair(key, value []byte, delete bool) (indexerbase.ObjectUpdate, bool, error) {
	// strip prefix
	key = key[len(c.GetPrefix()):]
	var k any
	var err error
	if decodeAny, ok := c.m.kc.(IndexableCodec); ok {
		k, err = decodeAny.DecodeIndexable(key)
	} else {
		_, k, err = c.m.kc.Decode(key)
	}
	if err != nil {
		return indexerbase.ObjectUpdate{
			TypeName: c.GetName(),
		}, false, err
	}

	if delete {
		return indexerbase.ObjectUpdate{
			TypeName: c.GetName(),
			Key:      k,
			Delete:   true,
		}, true, nil
	}

	var v any
	if decodeAny, ok := c.m.vc.(IndexableCodec); ok {
		v, err = decodeAny.DecodeIndexable(value)
	} else {
		v, err = c.m.vc.Decode(value)
	}
	if err != nil {
		return indexerbase.ObjectUpdate{
			TypeName: c.GetName(),
		}, false, err
	}

	return indexerbase.ObjectUpdate{
		TypeName: c.GetName(),
		Key:      k,
		Value:    v,
	}, true, nil
}

type IndexableCodec interface {
	SchemaFields() []indexerbase.Field
	DecodeIndexable([]byte) (any, error)
}
