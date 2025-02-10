package ormkv_test

import (
	"bytes"
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/internal/testutil"
)

func TestPrimaryKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		keyCodec := testutil.TestKeyCodecGen(0, 5).Draw(t, "keyCodec")
		pkCodec, err := ormkv.NewPrimaryKeyCodec(
			keyCodec.Codec.Prefix(),
			(&testpb.ExampleTable{}).ProtoReflect().Type(),
			keyCodec.Codec.GetFieldNames(),
			proto.UnmarshalOptions{},
		)
		assert.NilError(t, err)
		for i := 0; i < 100; i++ {
			a := testutil.GenA.Draw(t, fmt.Sprintf("a%d", i))
			key := keyCodec.Codec.GetKeyValues(a.ProtoReflect())
			pk1 := &ormkv.PrimaryKeyEntry{
				TableName: aFullName,
				Key:       key,
				Value:     a,
			}
			k, v, err := pkCodec.EncodeEntry(pk1)
			assert.NilError(t, err)

			k2, v2, err := pkCodec.EncodeKVFromMessage(a.ProtoReflect())
			assert.NilError(t, err)
			assert.Assert(t, bytes.Equal(k, k2))
			assert.Assert(t, bytes.Equal(v, v2))

			entry2, err := pkCodec.DecodeEntry(k, v)
			assert.NilError(t, err)
			pk2 := entry2.(*ormkv.PrimaryKeyEntry)
			assert.Equal(t, 0, pkCodec.CompareKeys(pk1.Key, pk2.Key))
			assert.DeepEqual(t, pk1.Value, pk2.Value, protocmp.Transform())

			idxFields, pk3, err := pkCodec.DecodeIndexKey(k, v)
			assert.NilError(t, err)
			assert.Equal(t, 0, pkCodec.CompareKeys(pk1.Key, pk3))
			assert.Equal(t, 0, pkCodec.CompareKeys(pk1.Key, idxFields))

			pkCodec.ClearValues(a.ProtoReflect())
			pkCodec.SetKeyValues(a.ProtoReflect(), pk1.Key)
			assert.DeepEqual(t, a, pk2.Value, protocmp.Transform())
		}
	})
}
