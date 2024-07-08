package collections

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tidwall/btree"

	"cosmossdk.io/collections/codec"
	"cosmossdk.io/schema"
)

// IndexingOptions are indexing options for the collections schema.
type IndexingOptions struct {

	// RetainDeletionsFor is the list of collections to retain deletions for.
	RetainDeletionsFor []string
}

// ModuleCodec returns the ModuleCodec for this schema for the provided options.
func (s Schema) ModuleCodec(opts IndexingOptions) (schema.ModuleCodec, error) {
	decoder := moduleDecoder{
		collectionLookup: &btree.Map[string, *collDecoder]{},
	}

	retainDeletions := make(map[string]bool)
	for _, collName := range opts.RetainDeletionsFor {
		retainDeletions[collName] = true
	}

	var objectTypes []schema.ObjectType
	for _, collName := range s.collectionsOrdered {
		coll := s.collectionsByName[collName]

		// skip secondary indexes
		if coll.isSecondaryIndex() {
			continue
		}

		ld, err := coll.schemaCodec()
		if err != nil {
			return schema.ModuleCodec{}, err
		}

		if retainDeletions[coll.GetName()] {
			ld.objectType.RetainDeletions = true
		}

		objectTypes = append(objectTypes, ld.objectType)

		decoder.collectionLookup.Set(string(coll.GetPrefix()), &collDecoder{
			Collection:  coll,
			schemaCodec: ld,
		})
	}

	return schema.ModuleCodec{
		Schema: schema.ModuleSchema{
			ObjectTypes: objectTypes,
		},
		KVDecoder: decoder.decodeKV,
	}, nil
}

type moduleDecoder struct {
	// collectionLookup lets us efficiently look the correct collection based on raw key bytes
	collectionLookup *btree.Map[string, *collDecoder]
}

func (m moduleDecoder) decodeKV(update schema.KVPairUpdate) ([]schema.ObjectUpdate, error) {
	key := update.Key
	ks := string(key)
	var cd *collDecoder
	// we look for the collection whose prefix is less than this key
	m.collectionLookup.Descend(ks, func(prefix string, cur *collDecoder) bool {
		bytesPrefix := cur.GetPrefix()
		if bytes.HasPrefix(key, bytesPrefix) {
			cd = cur
			return true
		}
		return false
	})
	if cd == nil {
		return nil, nil
	}

	return cd.decodeKVPair(update)
}

type collDecoder struct {
	Collection
	schemaCodec
}

func (c collDecoder) decodeKVPair(update schema.KVPairUpdate) ([]schema.ObjectUpdate, error) {
	// strip prefix
	key := update.Key
	key = key[len(c.GetPrefix()):]

	k, err := c.keyDecoder(key)
	if err != nil {
		return []schema.ObjectUpdate{
			{TypeName: c.GetName()},
		}, err

	}

	if update.Delete {
		return []schema.ObjectUpdate{
			{TypeName: c.GetName(), Key: k, Delete: true},
		}, nil
	}

	v, err := c.valueDecoder(update.Value)
	if err != nil {
		return []schema.ObjectUpdate{
			{TypeName: c.GetName(), Key: k},
		}, err
	}

	return []schema.ObjectUpdate{
		{TypeName: c.GetName(), Key: k, Value: v},
	}, nil
}

func (c collectionImpl[K, V]) schemaCodec() (schemaCodec, error) {
	res := schemaCodec{}
	res.objectType.Name = c.GetName()

	keyDecoder, err := codec.KeySchemaCodec(c.m.kc)
	if err != nil {
		return schemaCodec{}, err
	}
	res.objectType.KeyFields = keyDecoder.Fields
	res.keyDecoder = func(i []byte) (any, error) {
		_, x, err := c.m.kc.Decode(i)
		if err != nil {
			return nil, err
		}
		return keyDecoder.ToSchemaType(x)
	}
	ensureFieldNames(c.m.kc, "key", res.objectType.KeyFields)

	valueDecoder, err := codec.ValueSchemaCodec(c.m.vc)
	if err != nil {
		return schemaCodec{}, err
	}
	res.objectType.ValueFields = valueDecoder.Fields
	res.valueDecoder = func(i []byte) (any, error) {
		x, err := c.m.vc.Decode(i)
		if err != nil {
			return nil, err
		}
		return valueDecoder.ToSchemaType(x)
	}
	ensureFieldNames(c.m.vc, "value", res.objectType.ValueFields)

	return res, nil
}

// ensureFieldNames makes sure that all fields have valid names - either the
// names were specified by user or they get filled in with defaults here
func ensureFieldNames(x any, defaultName string, cols []schema.Field) {
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
