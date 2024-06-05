package collections

import (
	"github.com/tidwall/btree"

	"cosmossdk.io/collections/codec"
	indexerbase "cosmossdk.io/indexer/base"
)

type decoder struct {
}

type DecoderOptions struct {
}

func NewDecoder(opts DecoderOptions) indexerbase.Decoder {
	return decoder{}
}

func (d decoder) ExtractModuleDecoder(moduleName string, module any) indexerbase.ModuleStateDecoder {
	hasCollections, ok := module.(HasCollections)
	if !ok {
		return nil
	}

	// TODO: create collections index

	return &moduleStateDecoder{
		schema: hasCollections.CollectionsSchema(),
	}
}

var _ indexerbase.Decoder = &decoder{}

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

func (m moduleStateDecoder) DecodeSet(key, value []byte) (indexerbase.EntityUpdate, bool, error) {
	coll := m.getCollectionForKey(key)
	if coll == nil {
		return indexerbase.EntityUpdate{}, false, nil
	}

	return coll.decodeSet(key, value)
}

func (m moduleStateDecoder) DecodeDelete(key []byte) (indexerbase.EntityDelete, bool, error) {
	coll := m.getCollectionForKey(key)
	return coll.decodeDelete(key)
}

func (c collectionImpl[K, V]) getTableSchema() indexerbase.Table {
	var keyFields []indexerbase.Column
	var valueFields []indexerbase.Column

	if hasSchema, ok := c.m.kc.(codec.HasSchema); ok {
		keyFields = hasSchema.Fields()
	} else {
		name := "key"
		if named, ok := c.m.kc.(codec.HasName); ok {
			name = named.Name()
		}
		var k K
		keyFields, _ = extractFields(k, name)
	}

	if hasSchema, ok := c.m.vc.(codec.HasSchema); ok {
		valueFields = hasSchema.Fields()
	} else {
		name := "key"
		if named, ok := c.m.vc.(codec.HasName); ok {
			name = named.Name()
		}
		var v V
		valueFields, _ = extractFields(v, name)
	}

	return indexerbase.Table{
		Name:         c.GetName(),
		KeyColumns:   keyFields,
		ValueColumns: valueFields,
	}
}

func extractFields(x any, name string) ([]indexerbase.Column, func(any) any) {
	switch x.(type) {
	case int8:
	}
	panic("TODO")
}

func (c collectionImpl[K, V]) decodeSet(key, value []byte) (indexerbase.EntityUpdate, bool, error) {
	_, _, err := c.m.kc.Decode(key)
	if err != nil {
		return indexerbase.EntityUpdate{}, false, err
	}

	_, err = c.m.vc.Decode(value)
	if err != nil {
		return indexerbase.EntityUpdate{}, false, err
	}

	panic("TODO")
}

func (c collectionImpl[K, V]) decodeDelete(key []byte) (indexerbase.EntityDelete, bool, error) {
	panic("TODO")
}
