package ormkv_test

import (
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
		tableName := (&testpb.A{}).ProtoReflect().Descriptor().FullName()
		fields := ormkv.FieldsFromDescriptors(keyCodec.Codec.GetFieldDescriptors())
		uniqueKeyCdc, err := ormkv.NewUniqueKeyCodec(keyCodec.Codec, tableName, pkCodec.Codec.GetFieldDescriptors())
		assert.NilError(t, err)
		for i := 0; i < 100; i++ {
			key := keyCodec.Draw(t, fmt.Sprintf("key%d", i))
			pk := keyCodec.Draw(t, fmt.Sprintf("pk%d", i))
			uniq1 := ormkv.IndexKeyEntry{
				TableName:   tableName,
				Fields:      fields,
				IsUnique:    true,
				IndexValues: key,
				PrimaryKey:  pk,
			}
			k, v, err := uniqueKeyCdc.EncodeKV(uniq1)
			assert.NilError(t, err)

			entry2, err := uniqueKeyCdc.DecodeKV(k, v)
			assert.NilError(t, err)
			uniq2 := entry2.(ormkv.IndexKeyEntry)
			assert.Equal(t, 0, keyCodec.Codec.CompareValues(uniq1.IndexValues, uniq2.IndexValues))
			assert.Equal(t, 0, pkCodec.Codec.CompareValues(uniq1.PrimaryKey, uniq2.PrimaryKey))

			idxFields, pk2, err := uniqueKeyCdc.DecodeIndexKey(k, v)
			assert.NilError(t, err)
			assert.Equal(t, 0, keyCodec.Codec.CompareValues(key, idxFields))
			assert.Equal(t, 0, pkCodec.Codec.CompareValues(key, pk2))
		}
	})
}
