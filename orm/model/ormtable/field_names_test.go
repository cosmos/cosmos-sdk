package ormtable

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"

	"gotest.tools/v3/assert"
)

func TestFieldNames(t *testing.T) {
	names := []protoreflect.Name{"a", "b", "c"}

	abc := "a,b,c"
	f, err := CommaSeparatedFieldNames(abc)
	assert.NilError(t, err)
	assert.Equal(t, FieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	assert.DeepEqual(t, names, f.Names())

	f, err = CommaSeparatedFieldNames("a, b ,c")
	assert.NilError(t, err)
	assert.Equal(t, FieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	// empty okay
	f, err = CommaSeparatedFieldNames("")
	assert.NilError(t, err)
	assert.Equal(t, FieldNames{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())

	f = FieldsFromNames(names)
	assert.Equal(t, FieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	// empty okay
	f = FieldsFromNames([]protoreflect.Name{})
	assert.Equal(t, FieldNames{""}, f)
	f = FieldsFromNames(nil)
	assert.Equal(t, FieldNames{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())

	aFields := (&testpb.A{}).ProtoReflect().Descriptor().Fields()
	f = FieldsFromDescriptors([]protoreflect.FieldDescriptor{
		aFields.ByName("u32"),
		aFields.ByName("e"),
	})
	assert.Equal(t, FieldNames{"u32,e"}, f)
	assert.Equal(t, "u32,e", f.String())
	assert.DeepEqual(t, []protoreflect.Name{"u32", "e"}, f.Names())

	// empty okay
	f = FieldsFromDescriptors([]protoreflect.FieldDescriptor{})
	assert.Equal(t, FieldNames{""}, f)
	f = FieldsFromDescriptors(nil)
	assert.Equal(t, FieldNames{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())
}
