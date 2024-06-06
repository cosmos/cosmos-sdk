package collections

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tidwall/btree"

	indexerbase "cosmossdk.io/indexer/base"
)

type IndexingOptions struct {
}

func (s Schema) ModuleDecoder(opts IndexingOptions) (indexerbase.ModuleDecoder, error) {
	decoder := moduleDecoder{
		lookup: &btree.Map[string, *collDecoder]{},
	}

	var tables []indexerbase.Table
	for _, collName := range s.collectionsOrdered {
		coll := s.collectionsByName[collName]
		if coll.isIndex() {
			continue
		}

		schema := coll.getTableSchema()
		tables = append(tables, schema)
		decoder.lookup.Set(string(coll.GetPrefix()), &collDecoder{Collection: coll})
	}
	return indexerbase.ModuleDecoder{
		Schema: indexerbase.Schema{
			Tables: tables,
		},
		KVDecoder: decoder.decodeKV,
	}, nil
}

type moduleDecoder struct {
	lookup *btree.Map[string, *collDecoder]
}

func (m moduleDecoder) decodeKV(key, value []byte) (indexerbase.EntityUpdate, bool, error) {
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
		return indexerbase.EntityUpdate{}, false, nil
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

func (c collectionImpl[K, V]) getTableSchema() indexerbase.Table {
	var keyFields []indexerbase.Column
	var valueFields []indexerbase.Column

	if hasSchema, ok := c.m.kc.(HasSchema); ok {
		keyFields = hasSchema.SchemaColumns()
	} else {
		var k K
		keyFields, _ = extractFields(k)
	}
	ensureNames(c.m.kc, "key", keyFields)

	if hasSchema, ok := c.m.vc.(HasSchema); ok {
		valueFields = hasSchema.SchemaColumns()
	} else {
		var v V
		valueFields, _ = extractFields(v)
	}
	ensureNames(c.m.vc, "value", valueFields)

	return indexerbase.Table{
		Name:         c.GetName(),
		KeyColumns:   keyFields,
		ValueColumns: valueFields,
	}
}

func extractFields(x any) ([]indexerbase.Column, func(any) any) {
	if hasSchema, ok := x.(HasSchema); ok {
		return hasSchema.SchemaColumns(), nil
	}

	ty := indexerbase.TypeForGoValue(x)
	if ty >= 0 {
		return []indexerbase.Column{{Type: ty}}, nil
	}

	if _, ok := x.(interface{ String() string }); ok {
		return []indexerbase.Column{
			{
				Type: indexerbase.TypeString,
			},
		}, func(x any) any { return x.(interface{ String() string }).String() }
	}

	panic(fmt.Errorf("unsupported type %T", x))
}

func ensureNames(x any, defaultName string, cols []indexerbase.Column) {
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

func (c collectionImpl[K, V]) decodeKVPair(key, value []byte, delete bool) (indexerbase.EntityUpdate, bool, error) {
	// strip prefix
	key = key[len(c.GetPrefix()):]
	_, k, err := c.m.kc.Decode(key)
	if err != nil {
		return indexerbase.EntityUpdate{}, false, err
	}

	if delete {
		return indexerbase.EntityUpdate{
			TableName: c.GetName(),
			Key:       k,
			Delete:    true,
		}, true, nil
	}

	v, err := c.m.vc.Decode(value)
	if err != nil {
		return indexerbase.EntityUpdate{
			TableName: c.GetName(),
		}, false, err
	}

	return indexerbase.EntityUpdate{
		TableName: c.GetName(),
		Key:       v,
		Delete:    true,
	}, true, nil
}

type HasSchema interface {
	SchemaColumns() []indexerbase.Column
}

type DecodeAny interface {
	DecodeAny([]byte) (any, error)
}
