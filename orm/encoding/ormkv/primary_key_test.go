package ormkv_test

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
)

func TestPrimaryKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		keyCodec := testutil.TestKeyCodecGen.Draw(t, "keyCodec").(testutil.TestKeyCodec)
		pkCodec := ormkv.NewPrimaryKeyCodec(keyCodec.Codec, (&testpb.A{}).ProtoReflect().Type(), proto.UnmarshalOptions{})
		for i := 0; i < 100; i++ {
			key := keyCodec.Draw(t, fmt.Sprintf("i%d", i))
			pk1 := ormkv.PrimaryKeyEntry{
				Key:   key,
				Value: &testpb.A{},
			}
			k, v, err := pkCodec.EncodeKV(pk1)
			assert.NilError(t, err)

			entry2, err := pkCodec.DecodeKV(k, v)
			assert.NilError(t, err)
			pk2 := entry2.(ormkv.PrimaryKeyEntry)
			assert.Equal(t, 0, pkCodec.CompareValues(pk1.Key, pk2.Key))
			assert.DeepEqual(t, pk1.Value, pk2.Value, protocmp.Transform())

			idxFields, pk3, err := pkCodec.DecodeIndexKey(k, v)
			assert.NilError(t, err)
			assert.Equal(t, 0, pkCodec.CompareValues(pk1.Key, pk3))
			assert.Equal(t, 0, pkCodec.CompareValues(pk1.Key, idxFields))

			var a1, a2 testpb.A
			pkCodec.SetValues(a1.ProtoReflect(), pk1.Key)
			err = pkCodec.Unmarshal(pk1.Key, v, &a2)
			assert.NilError(t, err)
			assert.DeepEqual(t, &a1, &a2, protocmp.Transform())
		}
	})
}
