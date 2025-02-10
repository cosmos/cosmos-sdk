package ormkv_test

import (
	"bytes"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/internal/testutil"
)

func TestIndexKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		idxPartCdc := testutil.TestKeyCodecGen(1, 5).Draw(t, "idxPartCdc")
		pkCodec := testutil.TestKeyCodecGen(1, 5).Draw(t, "pkCdc")
		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix")
		messageType := (&testpb.ExampleTable{}).ProtoReflect().Type()
		indexKeyCdc, err := ormkv.NewIndexKeyCodec(
			prefix,
			messageType,
			idxPartCdc.Codec.GetFieldNames(),
			pkCodec.Codec.GetFieldNames(),
		)
		assert.NilError(t, err)
		for i := 0; i < 100; i++ {
			a := testutil.GenA.Draw(t, fmt.Sprintf("a%d", i))
			key := indexKeyCdc.GetKeyValues(a.ProtoReflect())
			pk := pkCodec.Codec.GetKeyValues(a.ProtoReflect())
			idx1 := &ormkv.IndexKeyEntry{
				TableName:   messageType.Descriptor().FullName(),
				Fields:      indexKeyCdc.GetFieldNames(),
				IsUnique:    false,
				IndexValues: key,
				PrimaryKey:  pk,
			}
			k, v, err := indexKeyCdc.EncodeEntry(idx1)
			assert.NilError(t, err)

			k2, v2, err := indexKeyCdc.EncodeKVFromMessage(a.ProtoReflect())
			assert.NilError(t, err)
			assert.Assert(t, bytes.Equal(k, k2))
			assert.Assert(t, bytes.Equal(v, v2))

			entry2, err := indexKeyCdc.DecodeEntry(k, v)
			assert.NilError(t, err)
			idx2 := entry2.(*ormkv.IndexKeyEntry)
			assert.Equal(t, 0, indexKeyCdc.CompareKeys(idx1.IndexValues, idx2.IndexValues))
			assert.Equal(t, 0, pkCodec.Codec.CompareKeys(idx1.PrimaryKey, idx2.PrimaryKey))
			assert.Equal(t, false, idx2.IsUnique)
			assert.Equal(t, messageType.Descriptor().FullName(), idx2.TableName)
			assert.DeepEqual(t, idx1.Fields, idx2.Fields)

			idxFields, pk2, err := indexKeyCdc.DecodeIndexKey(k, v)
			assert.NilError(t, err)
			assert.Equal(t, 0, indexKeyCdc.CompareKeys(key, idxFields))
			assert.Equal(t, 0, pkCodec.Codec.CompareKeys(pk, pk2))
		}
	})
}
