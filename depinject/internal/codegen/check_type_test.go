package codegen

import (
	"reflect"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject/internal/graphviz"
)

func TestCheckIsExportedType(t *testing.T) {
	expectValidType(t, false)
	expectValidType(t, uint(0))
	expectValidType(t, uint8(0))
	expectValidType(t, uint16(0))
	expectValidType(t, uint32(0))
	expectValidType(t, uint64(0))
	expectValidType(t, int(0))
	expectValidType(t, int8(0))
	expectValidType(t, int16(0))
	expectValidType(t, int32(0))
	expectValidType(t, int64(0))
	expectValidType(t, float32(0))
	expectValidType(t, float64(0))
	expectValidType(t, complex64(0))
	expectValidType(t, complex128(0))
	expectValidType(t, MyInt(0))
	expectValidType(t, [1]int{0})
	expectValidType(t, []int{})
	expectValidType(t, make(chan int))
	expectValidType(t, make(<-chan int))
	expectValidType(t, make(chan<- int))
	expectValidType(t, func(int, string) (bool, error) { return false, nil })
	expectValidType(t, func(int, ...string) (bool, error) { return false, nil })
	expectValidType(t, AStruct{})
	expectValidType(t, map[string]graphviz.Attributes{})
	expectValidType(t, &AStruct{})
	expectValidType(t, AGenericStruct[graphviz.Node, FileGen]{})
	expectValidType(t, AStructWrapper{})
	expectValidType(t, "abc")
	expectValidType(t, uintptr(0))
	expectValidType(t, (*AnInterface)(nil))
}

func expectValidType(t *testing.T, v interface{}) {
	assert.NilError(t, CheckIsExportedType(reflect.TypeOf(v)))
}
