package collections

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"github.com/tidwall/btree"
	"google.golang.org/protobuf/reflect/protoreflect"

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
		collectionLookup: &btree.Map[string, *collectionSchemaCodec]{},
	}

	retainDeletions := make(map[string]bool)
	for _, collName := range opts.RetainDeletionsFor {
		retainDeletions[collName] = true
	}

	var types []schema.Type
	for _, collName := range s.collectionsOrdered {
		coll := s.collectionsByName[collName]

		// skip secondary indexes
		if coll.isSecondaryIndex() {
			continue
		}

		cdc, err := coll.schemaCodec()
		if err != nil {
			return schema.ModuleCodec{}, err
		}

		if retainDeletions[coll.GetName()] {
			cdc.objectType.RetainDeletions = true
		}

		// this part below is a bit hacky, it will try to convert to a proto.Message
		// in order to get any enum types inside of it.
		emptyVal, err := coll.ValueCodec().Decode([]byte{})
		if err == nil {
			// convert to proto.Message
			pt, err := toProtoMessage(emptyVal)
			if err == nil {
				msgName := proto.MessageName(pt)
				desc, err := proto.HybridResolver.FindDescriptorByName(protoreflect.FullName(msgName))
				if err != nil {
					return schema.ModuleCodec{}, fmt.Errorf("could not find descriptor for %s: %w", msgName, err)
				}
				msgDesc := desc.(protoreflect.MessageDescriptor)

				// go through enum descriptors and add them to types
				for i := 0; i < msgDesc.Fields().Len(); i++ {
					field := msgDesc.Fields().Get(i)
					enum := field.Enum()
					if enum == nil {
						continue
					}

					enumType := schema.EnumType{
						Name: strings.ReplaceAll(string(enum.FullName()), ".", "_"), // make it compatible with schema
					}
					for j := 0; j < enum.Values().Len(); j++ {
						val := enum.Values().Get(j)
						enumType.Values = append(enumType.Values, schema.EnumValueDefinition{
							Name:  string(val.Name()),
							Value: int32(val.Number()),
						})
					}
					types = append(types, enumType)
				}

			}
		}

		types = append(types, cdc.objectType)
		decoder.collectionLookup.Set(string(coll.GetPrefix()), cdc)
	}

	modSchema, err := schema.CompileModuleSchema(types...)
	if err != nil {
		return schema.ModuleCodec{}, err
	}

	return schema.ModuleCodec{
		Schema:    modSchema,
		KVDecoder: decoder.decodeKV,
	}, nil
}

type moduleDecoder struct {
	// collectionLookup lets us efficiently look the correct collection based on raw key bytes
	collectionLookup *btree.Map[string, *collectionSchemaCodec]
}

func (m moduleDecoder) decodeKV(update schema.KVPairUpdate) ([]schema.StateObjectUpdate, error) {
	key := update.Key
	ks := string(key)
	var cd *collectionSchemaCodec
	// we look for the collection whose prefix is less than this key
	m.collectionLookup.Descend(ks, func(prefix string, cur *collectionSchemaCodec) bool {
		bytesPrefix := cur.coll.GetPrefix()
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

func (c collectionSchemaCodec) decodeKVPair(update schema.KVPairUpdate) ([]schema.StateObjectUpdate, error) {
	// strip prefix
	key := update.Key
	key = key[len(c.coll.GetPrefix()):]

	k, err := c.keyDecoder(key)
	if err != nil {
		return []schema.StateObjectUpdate{
			{TypeName: c.coll.GetName()},
		}, err
	}

	if update.Remove {
		return []schema.StateObjectUpdate{
			{TypeName: c.coll.GetName(), Key: k, Delete: true},
		}, nil
	}

	v, err := c.valueDecoder(update.Value)
	if err != nil {
		return []schema.StateObjectUpdate{
			{TypeName: c.coll.GetName(), Key: k},
		}, err
	}

	return []schema.StateObjectUpdate{
		{TypeName: c.coll.GetName(), Key: k, Value: v},
	}, nil
}

func (c collectionImpl[K, V]) schemaCodec() (*collectionSchemaCodec, error) {
	res := &collectionSchemaCodec{
		coll: c,
	}
	res.objectType.Name = c.GetName()

	keyDecoder, err := codec.KeySchemaCodec(c.m.kc)
	if err != nil {
		return nil, err
	}
	res.objectType.KeyFields = keyDecoder.Fields
	res.keyDecoder = func(i []byte) (any, error) {
		_, x, err := c.m.kc.Decode(i)
		if err != nil {
			return nil, err
		}
		if keyDecoder.ToSchemaType == nil {
			return x, nil
		}
		return keyDecoder.ToSchemaType(x)
	}
	ensureFieldNames(c.m.kc, "key", res.objectType.KeyFields)

	valueDecoder, err := codec.ValueSchemaCodec(c.m.vc)
	if err != nil {
		return nil, err
	}
	res.objectType.ValueFields = valueDecoder.Fields
	res.valueDecoder = func(i []byte) (any, error) {
		x, err := c.m.vc.Decode(i)
		if err != nil {
			return nil, err
		}

		if valueDecoder.ToSchemaType == nil {
			return x, nil
		}

		return valueDecoder.ToSchemaType(x)
	}
	ensureFieldNames(c.m.vc, "value", res.objectType.ValueFields)

	return res, nil
}

// ensureFieldNames makes sure that all fields have valid names - either the
// names were specified by user or they get filled
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
		} else if col.Name == "" {
			if i == 0 && len(cols) == 1 {
				col.Name = defaultName
			} else {
				col.Name = fmt.Sprintf("%s%d", defaultName, i+1)
			}
		}
		cols[i] = col
	}
}

// toProtoMessage is a helper to convert a value to a proto.Message.
func toProtoMessage(value interface{}) (proto.Message, error) {
	if value == nil {
		return nil, fmt.Errorf("value is nil")
	}

	// Check if the value already implements proto.Message
	if msg, ok := value.(proto.Message); ok {
		return msg, nil
	}

	// Use reflection to handle non-pointer values
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		// Already a pointer, but doesn't implement proto.Message
		return nil, fmt.Errorf("value is a pointer but does not implement proto.Message")
	}

	// If not a pointer, create a pointer to the value dynamically
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)

	// Assert if the pointer implements proto.Message
	msg, ok := ptr.Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("value does not implement proto.Message")
	}

	return msg, nil
}
