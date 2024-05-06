package ormkv_test

import (
	"bytes"
	"fmt"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/internal/testutil"
	"cosmossdk.io/orm/types/ormerrors"
)

func TestUniqueKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		keyCodec := testutil.TestKeyCodecGen(1, 5).Draw(t, "keyCodec")
		pkCodec := testutil.TestKeyCodecGen(1, 5).Draw(t, "primaryKeyCodec")

		// check if we have a trivial unique index where all of the fields
		// in the primary key are in the unique key, we should expect an
		// error in this case
		isInPk := map[protoreflect.Name]bool{}
		for _, spec := range pkCodec.KeySpecs {
			isInPk[spec.FieldName] = true
		}
		numPkFields := 0
		for _, spec := range keyCodec.KeySpecs {
			if isInPk[spec.FieldName] {
				numPkFields++
			}
		}
		isTrivialUniqueKey := numPkFields == len(pkCodec.KeySpecs)

		messageType := (&testpb.ExampleTable{}).ProtoReflect().Type()
		uniqueKeyCdc, err := ormkv.NewUniqueKeyCodec(
			keyCodec.Codec.Prefix(),
			messageType,
			keyCodec.Codec.GetFieldNames(),
			pkCodec.Codec.GetFieldNames(),
		)

		if isTrivialUniqueKey {
			assert.ErrorContains(t, err, "no new uniqueness constraint")
			return
		}
		assert.NilError(t, err)

		for i := 0; i < 100; i++ {
			a := testutil.GenA.Draw(t, fmt.Sprintf("a%d", i))
			key := keyCodec.Codec.GetKeyValues(a.ProtoReflect())
			pk := pkCodec.Codec.GetKeyValues(a.ProtoReflect())
			uniq1 := &ormkv.IndexKeyEntry{
				TableName:   messageType.Descriptor().FullName(),
				Fields:      keyCodec.Codec.GetFieldNames(),
				IsUnique:    true,
				IndexValues: key,
				PrimaryKey:  pk,
			}
			k, v, err := uniqueKeyCdc.EncodeEntry(uniq1)
			assert.NilError(t, err)

			k2, v2, err := uniqueKeyCdc.EncodeKVFromMessage(a.ProtoReflect())
			assert.NilError(t, err)
			assert.Assert(t, bytes.Equal(k, k2))
			assert.Assert(t, bytes.Equal(v, v2))

			entry2, err := uniqueKeyCdc.DecodeEntry(k, v)
			assert.NilError(t, err)
			uniq2 := entry2.(*ormkv.IndexKeyEntry)
			assert.Equal(t, 0, keyCodec.Codec.CompareKeys(uniq1.IndexValues, uniq2.IndexValues))
			assert.Equal(t, 0, pkCodec.Codec.CompareKeys(uniq1.PrimaryKey, uniq2.PrimaryKey))
			assert.Equal(t, true, uniq2.IsUnique)
			assert.Equal(t, messageType.Descriptor().FullName(), uniq2.TableName)
			assert.DeepEqual(t, uniq1.Fields, uniq2.Fields)

			idxFields, pk2, err := uniqueKeyCdc.DecodeIndexKey(k, v)
			assert.NilError(t, err)
			assert.Equal(t, 0, keyCodec.Codec.CompareKeys(key, idxFields))
			assert.Equal(t, 0, pkCodec.Codec.CompareKeys(pk, pk2))
		}
	})
}

func TestTrivialUnique(t *testing.T) {
	_, err := ormkv.NewUniqueKeyCodec(nil, (&testpb.ExampleTable{}).ProtoReflect().Type(),
		[]protoreflect.Name{"u32", "str"}, []protoreflect.Name{"str", "u32"})
	assert.ErrorIs(t, err, ormerrors.InvalidTableDefinition)
}
