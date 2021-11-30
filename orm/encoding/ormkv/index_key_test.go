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

func TestIndexKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		idxPartCdc := testutil.TestKeyCodecGen(1, 5).Draw(t, "idxPartCdc").(testutil.TestKeyCodec)
		pkCodec := testutil.TestKeyCodecGen(1, 5).Draw(t, "pkCdc").(testutil.TestKeyCodec)
		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)
		tableName := (&testpb.A{}).ProtoReflect().Descriptor().FullName()
		indexKeyCdc, err := ormkv.NewIndexKeyCodec(
			prefix,
			tableName,
			idxPartCdc.Codec.GetFieldDescriptors(),
			pkCodec.Codec.GetFieldDescriptors(),
		)
		assert.NilError(t, err)
		fields := ormkv.FieldsFromDescriptors(indexKeyCdc.GetFieldDescriptors())
		for i := 0; i < 100; i++ {
			a := testutil.GenA.Draw(t, fmt.Sprintf("a%d", i)).(*testpb.A)
			key := indexKeyCdc.GetValues(a.ProtoReflect())
			pk := pkCodec.Codec.GetValues(a.ProtoReflect())
			idx1 := &ormkv.IndexKeyEntry{
				TableName:   tableName,
				Fields:      fields,
				IsUnique:    false,
				IndexValues: key,
				PrimaryKey:  pk,
			}
			k, v, err := indexKeyCdc.EncodeEntry(idx1)
			assert.NilError(t, err)

			entry2, err := indexKeyCdc.DecodeEntry(k, v)
			assert.NilError(t, err)
			idx2 := entry2.(*ormkv.IndexKeyEntry)
			assert.Equal(t, 0, indexKeyCdc.CompareValues(idx1.IndexValues, idx2.IndexValues))
			assert.Equal(t, 0, pkCodec.Codec.CompareValues(idx1.PrimaryKey, idx2.PrimaryKey))
			assert.Equal(t, false, idx2.IsUnique)
			assert.Equal(t, tableName, idx2.TableName)
			assert.Equal(t, idx1.Fields, idx2.Fields)

			idxFields, pk2, err := indexKeyCdc.DecodeIndexKey(k, v)
			assert.NilError(t, err)
			assert.Equal(t, 0, indexKeyCdc.CompareValues(key, idxFields))
			assert.Equal(t, 0, pkCodec.Codec.CompareValues(pk, pk2))
		}
	})
}
