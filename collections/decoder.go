package collections

import (
	"fmt"

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
	tables := make([]indexerbase.Table, 0, len(m.schema.collectionsByName))
	for _, collName := range m.schema.collectionsOrdered {
		coll := m.schema.collectionsByName[collName]
		if coll.isIndex() {
			continue
		}

		schema := coll.getTableSchema()
		tables = append(tables, schema)
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
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64, string:
		ty, f := simpleType(x)
		return []indexerbase.Column{
			{
				Name: name,
				Type: ty,
			},
		}, f
	default:
		panic(fmt.Errorf("unsupported type: %T", x))
	}
}

func simpleType(x any) (indexerbase.Type, func(any) any) {
	switch x.(type) {
	case int8:
		return indexerbase.TypeInt8, func(x any) any { return x.(int8) }
	case int16:
		return indexerbase.TypeInt16, func(x any) any { return x.(int16) }
	case int32:
		return indexerbase.TypeInt32, func(x any) any { return x.(int32) }
	case int64:
		return indexerbase.TypeInt64, func(x any) any { return x.(int64) }
	case string:
		return indexerbase.TypeString, func(x any) any { return x.(string) }
	case []byte:
		return indexerbase.TypeBytes, func(x any) any { return x.([]byte) }
	case bool:
		return indexerbase.TypeBool, func(x any) any { return x.(bool) }
	case float32:
		return indexerbase.TypeFloat32, func(x any) any { return x.(float32) }
	case float64:
		return indexerbase.TypeFloat64, func(x any) any { return x.(float64) }
	case uint8:
		return indexerbase.TypeInt16, func(x any) any { return int16(x.(uint8)) }
	case uint16:
		return indexerbase.TypeInt32, func(x any) any { return int32(x.(uint16)) }
	case uint32:
		return indexerbase.TypeInt64, func(x any) any { return int64(x.(uint32)) }
	case uint64:
		return indexerbase.TypeDecimal, func(x any) any { panic("TODO") }
	default:
		panic(fmt.Errorf("unsupported type: %T", x))
	}
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
