package ormkv

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"

	"gotest.tools/v3/assert"
)

func TestFields(t *testing.T) {
	names := []protoreflect.Name{"a", "b", "c"}

	abc := "a,b,c"
	f, err := CommaSeparatedFields(abc)
	assert.NilError(t, err)
	assert.Equal(t, Fields{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	assert.DeepEqual(t, names, f.Names())

	f, err = CommaSeparatedFields("a, b ,c")
	assert.NilError(t, err)
	assert.Equal(t, Fields{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	// empty okay
	f, err = CommaSeparatedFields("")
	assert.NilError(t, err)
	assert.Equal(t, Fields{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())

	f = FieldsFromNames(names)
	assert.Equal(t, Fields{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	// empty okay
	f = FieldsFromNames([]protoreflect.Name{})
	assert.Equal(t, Fields{""}, f)
	f = FieldsFromNames(nil)
	assert.Equal(t, Fields{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())

	aFields := (&testpb.A{}).ProtoReflect().Descriptor().Fields()
	f = FieldsFromDescriptors([]protoreflect.FieldDescriptor{
		aFields.ByName("u32"),
		aFields.ByName("e"),
	})
	assert.Equal(t, Fields{"u32,e"}, f)
	assert.Equal(t, "u32,e", f.String())
	assert.DeepEqual(t, []protoreflect.Name{"u32", "e"}, f.Names())

	// empty okay
	f = FieldsFromDescriptors([]protoreflect.FieldDescriptor{})
	assert.Equal(t, Fields{""}, f)
	f = FieldsFromDescriptors(nil)
	assert.Equal(t, Fields{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())
}
