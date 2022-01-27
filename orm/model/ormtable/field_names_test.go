package ormtable

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"gotest.tools/v3/assert"
)

func TestFieldNames(t *testing.T) {
	names := []protoreflect.Name{"a", "b", "c"}

	abc := "a,b,c"
	f := commaSeparatedFieldNames(abc)
	assert.Equal(t, fieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	f = commaSeparatedFieldNames("a, b ,c")
	assert.Equal(t, fieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	// empty okay
	f = commaSeparatedFieldNames("")
	assert.Equal(t, fieldNames{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())

	f = fieldsFromNames(names)
	assert.Equal(t, fieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	// empty okay
	f = fieldsFromNames([]protoreflect.Name{})
	assert.Equal(t, fieldNames{""}, f)
	f = fieldsFromNames(nil)
	assert.Equal(t, fieldNames{""}, f)
	assert.Equal(t, 0, len(f.Names()))
	assert.Equal(t, "", f.String())
}
