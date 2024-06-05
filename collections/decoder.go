package collections

import (
	"github.com/tidwall/btree"

	"cosmossdk.io/collections/codec"
	indexerbase "cosmossdk.io/indexer/base"
)

type Decoder struct {
}

func (d Decoder) ExtractModuleDecoder(moduleName string, module any) indexerbase.ModuleStateDecoder {
	hasCollections, ok := module.(HasCollections)
	if !ok {
		return nil
	}

	// TODO: create collections index

	return &moduleStateDecoder{
		schema: hasCollections.CollectionsSchema(),
	}
}

var _ indexerbase.Decoder = &Decoder{}

type moduleStateDecoder struct {
	schema           Schema
	collectionsIndex btree.BTree
}

func (m moduleStateDecoder) GetSchema() indexerbase.Schema {
	tables := make([]indexerbase.Table, len(m.schema.collectionsByName))
	for _, collName := range m.schema.collectionsOrdered {
		coll := m.schema.collectionsByName[collName]
		tables = append(tables, coll.getTableSchema())
	}
	return indexerbase.Schema{
		Tables: tables,
	}
}

func (m moduleStateDecoder) getCollectionForKey(key []byte) Collection {
	panic("implement me")
}

func (m moduleStateDecoder) DecodeSet(key, value []byte) (indexerbase.EntityUpdate, error) {
	coll := m.getCollectionForKey(key)
	return coll.decodeSet(key, value)
}

func (m moduleStateDecoder) DecodeDelete(key []byte) (indexerbase.EntityDelete, error) {
	coll := m.getCollectionForKey(key)
	return coll.decodeDelete(key)
}

func (c collectionImpl[K, V]) getTableSchema() indexerbase.Table {
	var keyFields []indexerbase.Field
	var valueFields []indexerbase.Field

	if hasSchema, ok := c.m.kc.(codec.HasSchema); ok {
		keyFields = hasSchema.Fields()
	} else {
		name := "key"
		if named, ok := c.m.kc.(codec.HasName); ok {
			name = named.Name()
		}
		var k K
		keyFields = extractFields(k, name)
	}

	if hasSchema, ok := c.m.vc.(codec.HasSchema); ok {
		valueFields = hasSchema.Fields()
	} else {
		name := "key"
		if named, ok := c.m.vc.(codec.HasName); ok {
			name = named.Name()
		}
		var v V
		valueFields = extractFields(v, name)
	}

	return indexerbase.Table{
		Name:        c.GetName(),
		KeyFields:   keyFields,
		ValueFields: valueFields,
	}
}

func extractFields(x any, name string) []indexerbase.Field {
	switch x.(type) {
	case uint8, uint16, uint32:
		return []indexerbase.Field{{Name: name, Type: indexerbase.TypeUInt32}}
	case int8, int16, int32:
		return []indexerbase.Field{{Name: name, Type: indexerbase.TypeInt32}}
	case uint64:
		return []indexerbase.Field{{Name: name, Type: indexerbase.TypeUInt64}}
	}
	panic("TODO")
}

func (c collectionImpl[K, V]) decodeSet(key, value []byte) (indexerbase.EntityUpdate, error) {
	_, _, err := c.m.kc.Decode(key)
	if err != nil {
		return nil, err
	}

	_, err = c.m.vc.Decode(value)
	if err != nil {
		return nil, err
	}

	panic("TODO")
}

func (c collectionImpl[K, V]) decodeDelete(key []byte) (indexerbase.EntityDelete, error) {
	panic("TODO")
}
