package ormkv_test

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/encoding/encodeutil"
	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/internal/testpb"
)

var aFullName = (&testpb.ExampleTable{}).ProtoReflect().Descriptor().FullName()

func TestPrimaryKeyEntry(t *testing.T) {
	entry := &ormkv.PrimaryKeyEntry{
		TableName: aFullName,
		Key:       encodeutil.ValuesOf(uint32(1), "abc"),
		Value:     &testpb.ExampleTable{I32: -1},
	}
	assert.Equal(t, `PK testpb.ExampleTable 1/abc -> {"i32":-1}`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	// prefix key
	entry = &ormkv.PrimaryKeyEntry{
		TableName: aFullName,
		Key:       encodeutil.ValuesOf(uint32(1), "abc"),
		Value:     nil,
	}
	assert.Equal(t, `PK testpb.ExampleTable 1/abc -> _`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())
}

func TestIndexKeyEntry(t *testing.T) {
	entry := &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"u32", "i32", "str"},
		IsUnique:    false,
		IndexValues: encodeutil.ValuesOf(uint32(10), int32(-1), "abc"),
		PrimaryKey:  encodeutil.ValuesOf("abc", int32(-1)),
	}
	assert.Equal(t, `IDX testpb.ExampleTable u32/i32/str : 10/-1/abc -> abc/-1`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	entry = &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"u32"},
		IsUnique:    true,
		IndexValues: encodeutil.ValuesOf(uint32(10)),
		PrimaryKey:  encodeutil.ValuesOf("abc", int32(-1)),
	}
	assert.Equal(t, `UNIQ testpb.ExampleTable u32 : 10 -> abc/-1`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	// prefix key
	entry = &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"u32", "i32", "str"},
		IsUnique:    false,
		IndexValues: encodeutil.ValuesOf(uint32(10), int32(-1)),
	}
	assert.Equal(t, `IDX testpb.ExampleTable u32/i32/str : 10/-1 -> _`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())

	// prefix key
	entry = &ormkv.IndexKeyEntry{
		TableName:   aFullName,
		Fields:      []protoreflect.Name{"str", "i32"},
		IsUnique:    true,
		IndexValues: encodeutil.ValuesOf("abc", int32(1)),
	}
	assert.Equal(t, `UNIQ testpb.ExampleTable str/i32 : abc/1 -> _`, entry.String())
	assert.Equal(t, aFullName, entry.GetTableName())
}
