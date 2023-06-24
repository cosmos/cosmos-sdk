package fieldnames

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
)

func TestFieldNames(t *testing.T) {
	names := []protoreflect.Name{"a", "b", "c"}

	abc := "a,b,c"
	f := CommaSeparatedFieldNames(abc)
	assert.Equal(t, FieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	f = CommaSeparatedFieldNames("a, b ,c")
	assert.Equal(t, FieldNames{abc}, f)
	assert.DeepEqual(t, names, f.Names())
	assert.Equal(t, abc, f.String())

	// empty okay
	f = CommaSeparatedFieldNames("")
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
}
