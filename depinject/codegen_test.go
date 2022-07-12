package depinject

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"reflect"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/depinject/internal/graphviz"
)

func TestValueExpr(t *testing.T) {
	// bool
	expectValueExpr(t, true, `true`)
	expectValueExpr(t, false, `false`)

	// uints
	expectValueExpr(t, uint(0), `0`)
	expectValueExpr(t, uint8(1), `1`)
	expectValueExpr(t, uint16(2), `2`)
	expectValueExpr(t, uint32(3), `3`)
	expectValueExpr(t, uint64(12345678), `12345678`)

	// ints
	expectValueExpr(t, 0, `0`)
	expectValueExpr(t, int8(-1), `-1`)
	expectValueExpr(t, int16(-2), `-2`)
	expectValueExpr(t, int32(-3), `-3`)
	expectValueExpr(t, int64(-12345678), `-12345678`)

	// floats
	expectValueExpr(t, float32(0.0), `0e+00`)
	expectValueExpr(t, float64(1.32e-9), `1.32e-09`)

	// complex
	expectValueExpr(t, complex64(1+2i), `(1e+00+2e+00i)`)
	expectValueExpr(t, complex128(1.32e-9+-3.03i), `(1.32e-09-3.03e+00i)`)

	// array
	expectValueExpr(t, [3]uint32{1, 4, 9}, `[3]uint32{1, 4, 9}`)

	// slice
	expectValueExpr(t, []uint32{1, 4, 9}, `[]uint32{1, 4, 9}`)

	// map
	expectValueExpr(t, map[string]int{"a": 1}, `map[string]int{"a": 1}`)

	// struct
	expectValueExpr(t, AStruct{Foo: 2}, `depinject.AStruct{Foo: 2}`)
	expectValueExpr(t, AStruct{}, `depinject.AStruct{}`) // empty default fields

	// struct pointer
	expectValueExpr(t, &AStruct{Foo: 2}, `&depinject.AStruct{Foo: 2}`)
	var nilStruct *AStruct
	expectValueExpr(t, nilStruct, `nil`)

	// string
	expectValueExpr(t, "abc", `"abc"`)
}

func expectValueExpr(t *testing.T, value interface{}, expected string) {
	ctr := &container{}
	e, err := ctr.valueExpr(reflect.ValueOf(value))
	assert.NilError(t, err)
	expectExpr(t, e, expected)
}

type MyInt int

type AnInterface interface{}

type AStruct struct {
	Foo int
}

func TestTypeExpr(t *testing.T) {
	expectTypeExpr(t, false, "bool")
	expectTypeExpr(t, uint(0), "uint")
	expectTypeExpr(t, uint8(0), "uint8")
	expectTypeExpr(t, uint16(0), "uint16")
	expectTypeExpr(t, uint32(0), "uint32")
	expectTypeExpr(t, uint64(0), "uint64")
	expectTypeExpr(t, int(0), "int")
	expectTypeExpr(t, int8(0), "int8")
	expectTypeExpr(t, int16(0), "int16")
	expectTypeExpr(t, int32(0), "int32")
	expectTypeExpr(t, int64(0), "int64")
	expectTypeExpr(t, float32(0), "float32")
	expectTypeExpr(t, float64(0), "float64")
	expectTypeExpr(t, complex64(0), "complex64")
	expectTypeExpr(t, complex128(0), "complex128")
	expectTypeExpr(t, MyInt(0), "depinject.MyInt")
	expectTypeExpr(t, [1]int{0}, "[1]int")
	expectTypeExpr(t, []int{}, "[]int")
	expectTypeExpr(t, make(chan int), "chan int")
	expectTypeExpr(t, make(<-chan int), "<-chan int")
	expectTypeExpr(t, make(chan<- int), "chan<- int")
	expectTypeExpr(t, func(int, string) (bool, error) { return false, nil },
		"func(int, string) (bool, error)",
	)
	expectTypeExpr(t, func(int, ...string) (bool, error) { return false, nil },
		"func(int, ...string) (bool, error)",
	)
	expectTypeExpr(t, AStruct{}, "depinject.AStruct")
	expectTypeExpr(t, map[string]graphviz.Attributes{}, "map[string]graphviz.Attributes")
	expectTypeExpr(t, &AStruct{}, "*depinject.AStruct")
	expectTypeExpr(t, "abc", "string")
	expectTypeExpr(t, uintptr(0), "uintptr")
	// TODO: interface
	// TODO: UnsafePointer
}

func expectTypeExpr(t *testing.T, value interface{}, expected string) {
	ctr := &container{}
	e, err := ctr.typeExpr(reflect.TypeOf(value))
	assert.NilError(t, err)
	expectExpr(t, e, expected)
}

func expectExpr(t *testing.T, e ast.Expr, expected string) {
	fset := token.NewFileSet()
	buf := &bytes.Buffer{}
	assert.NilError(t, printer.Fprint(buf, fset, e))
	errBuf := &bytes.Buffer{}
	assert.NilError(t, ast.Fprint(errBuf, fset, e, nil))
	assert.Equal(t, expected, buf.String(), errBuf.String())
}
