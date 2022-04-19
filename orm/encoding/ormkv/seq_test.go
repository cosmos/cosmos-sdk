package ormkv_test

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"

	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
)

func TestSeqCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)
		typ := (&testpb.ExampleTable{}).ProtoReflect().Type()
		tableName := typ.Descriptor().FullName()
		cdc := ormkv.NewSeqCodec(typ, prefix)

		seq, err := cdc.DecodeValue(nil)
		assert.NilError(t, err)
		assert.Equal(t, uint64(0), seq)

		seq, err = cdc.DecodeValue([]byte{})
		assert.NilError(t, err)
		assert.Equal(t, uint64(0), seq)

		seq = rapid.Uint64().Draw(t, "seq").(uint64)

		v := cdc.EncodeValue(seq)
		seq2, err := cdc.DecodeValue(v)
		assert.NilError(t, err)
		assert.Equal(t, seq, seq2)

		entry := &ormkv.SeqEntry{
			TableName: tableName,
			Value:     seq,
		}
		k, v, err := cdc.EncodeEntry(entry)
		assert.NilError(t, err)
		entry2, err := cdc.DecodeEntry(k, v)
		assert.NilError(t, err)
		assert.DeepEqual(t, entry, entry2)
		assert.Assert(t, bytes.Equal(cdc.Prefix(), k))
	})
}
