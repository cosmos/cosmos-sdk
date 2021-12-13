package ormkv_test

import (
	"bytes"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
)

func TestUniqueKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		keyCodec := testutil.TestKeyCodecGen(1, 5).Draw(t, "keyCodec").(testutil.TestKeyCodec)
		pkCodec := testutil.TestKeyCodecGen(1, 5).Draw(t, "primaryKeyCodec").(testutil.TestKeyCodec)
		messageType := (&testpb.ExampleTable{}).ProtoReflect().Type()
		uniqueKeyCdc, err := ormkv.NewUniqueKeyCodec(
			keyCodec.Codec.Prefix(),
			messageType,
			keyCodec.Codec.GetFieldNames(),
			pkCodec.Codec.GetFieldNames(),
		)
		assert.NilError(t, err)
		for i := 0; i < 100; i++ {
			a := testutil.GenA.Draw(t, fmt.Sprintf("a%d", i)).(*testpb.ExampleTable)
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
