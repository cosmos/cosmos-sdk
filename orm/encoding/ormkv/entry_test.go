package ormkv_test

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
)

var aFullName = (&testpb.A{}).ProtoReflect().Descriptor().FullName()

func TestPrimaryKeyEntry(t *testing.T) {
	entry := &ormkv.PrimaryKeyEntry{
		TableName: aFullName,
		Key:       testutil.ValuesOf(uint32(1), "abc"),
		Value:     &testpb.A{I32: -1},
	}
	assert.Equal(t, `PK:testpb.A/1/"abc":{"i32":-1}`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	// prefix key
	entry = &ormkv.PrimaryKeyEntry{
		TableName: aFullName,
		Key:       testutil.ValuesOf(uint32(1), "abc"),
		Value:     nil,
	}
	assert.Equal(t, `PK:testpb.A/1/"abc":_`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())
}

func TestIndexKeyEntry(t *testing.T) {
	entry := &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"u32", "i32", "str"},
		IsUnique:    false,
		IndexValues: testutil.ValuesOf(uint32(10), int32(-1), "abc"),
		PrimaryKey:  testutil.ValuesOf("abc", int32(-1)),
	}
	assert.Equal(t, `IDX:testpb.A/u32/i32/str:10/-1/"abc":"abc"/-1`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	entry = &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"u32"},
		IsUnique:    true,
		IndexValues: testutil.ValuesOf(uint32(10)),
		PrimaryKey:  testutil.ValuesOf("abc", int32(-1)),
	}
	assert.Equal(t, `UNIQ:testpb.A/u32:10:"abc"/-1`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	// prefix key
	entry = &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"u32", "i32", "str"},
		IsUnique:    false,
		IndexValues: testutil.ValuesOf(uint32(10), int32(-1)),
	}
	assert.Equal(t, `IDX:testpb.A/u32/i32/str:10/-1:_`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	// prefix key
	entry = &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"str", "i32"},
		IsUnique:    true,
		IndexValues: testutil.ValuesOf("abc", int32(1)),
	}
	assert.Equal(t, `UNIQ:testpb.A/str/i32:"abc"/1:_`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())
}
