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

func TestIndexKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		idxPartCdc := testutil.TestKeyCodecGen(1, 5).Draw(t, "idxPartCdc").(testutil.TestKeyCodec)
		pkCodec := testutil.TestKeyCodecGen(1, 5).Draw(t, "pkCdc").(testutil.TestKeyCodec)
		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)
		desc := (&testpb.A{}).ProtoReflect().Descriptor()
		indexKeyCdc, err := ormkv.NewIndexKeyCodec(
			prefix,
			desc,
			idxPartCdc.Codec.GetFieldNames(),
			pkCodec.Codec.GetFieldNames(),
		)
		assert.NilError(t, err)
		for i := 0; i < 100; i++ {
			a := testutil.GenA.Draw(t, fmt.Sprintf("a%d", i)).(*testpb.A)
			key := indexKeyCdc.GetValues(a.ProtoReflect())
			pk := pkCodec.Codec.GetValues(a.ProtoReflect())
			idx1 := &ormkv.IndexKeyEntry{
				TableName:   desc.FullName(),
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
			assert.Equal(t, 0, indexKeyCdc.CompareValues(idx1.IndexValues, idx2.IndexValues))
			assert.Equal(t, 0, pkCodec.Codec.CompareValues(idx1.PrimaryKey, idx2.PrimaryKey))
			assert.Equal(t, false, idx2.IsUnique)
			assert.Equal(t, desc.FullName(), idx2.TableName)
			assert.DeepEqual(t, idx1.Fields, idx2.Fields)

			idxFields, pk2, err := indexKeyCdc.DecodeIndexKey(k, v)
			assert.NilError(t, err)
			assert.Equal(t, 0, indexKeyCdc.CompareValues(key, idxFields))
			assert.Equal(t, 0, pkCodec.Codec.CompareValues(pk, pk2))
		}
	})
}
