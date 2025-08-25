package depinject

import (
	"os"
	"reflect"
	"testing"

	"cosmossdk.io/depinject/internal/graphviz"
	"gotest.tools/v3/assert"
)

type testCase struct {
	name        string
	value       interface{}
	expectError string // "" means valid
}

func genTestCase() []testCase {
	return []testCase{
		// valid types
		{"Bool", false, ""},
		{"Uint", uint(0), ""},
		{"Uint8", uint8(0), ""},
		{"Uint16", uint16(0), ""},
		{"Uint32", uint32(0), ""},
		{"Uint64", uint64(0), ""},
		{"Int", int(0), ""},
		{"Int8", int8(0), ""},
		{"Int16", int16(0), ""},
		{"Int32", int32(0), ""},
		{"Int64", int64(0), ""},
		{"Float32", float32(0), ""},
		{"Float64", float64(0), ""},
		{"Complex64", complex64(0), ""},
		{"Complex128", complex128(0), ""},
		{"String", "", ""},
		{"OSFileMode", os.FileMode(0), ""},
		{"ArrayOfInt", [1]int{0}, ""},
		{"SliceOfInt", []int{}, ""},
		{"ChanOfInt", make(chan int), ""},
		{"RecvOnlyChan", make(<-chan int), ""},
		{"SendOnlyChan", make(chan<- int), ""},
		{"FunctionBasic", func(int, string) (bool, error) { return false, nil }, ""},
		{"FunctionVariadic", func(int, ...string) (bool, error) { return false, nil }, ""},
		{"ExportedStruct", In{}, ""},
		{"MapStringToExported", map[string]In{}, ""},
		{"PointerToExported", &In{}, ""},
		{"Uintptr", uintptr(0), ""},
		{"NilLocationPointer", (*Location)(nil), ""},

		// invalid types
		{"UnexportedStruct", container{}, "must be exported"},
		{"PointerToUnexportedStruct", &container{}, "must be exported"},
		{"InternalTypeGraphviz", graphviz.Attributes{}, "internal"},
		{"MapWithInternalType", map[string]graphviz.Attributes{}, "internal"},
		{"SliceWithInternalType", []graphviz.Attributes{}, "internal"},
	}
}

func TestIsExportedType(t *testing.T) {
	cases := genTestCase()

	for _, tc := range cases {
		rv := reflect.TypeOf(tc.value)
		t.Run(tc.name, func(t *testing.T) {
			if err := isExportedType(rv); err != nil {
				assert.ErrorContains(t, err, tc.expectError)
			} else {
				if tc.expectError != "" {
					t.FailNow()
				}
			}
		})
	}
}

func BenchmarkIsExportedType(b *testing.B) {
	cases := genTestCase()
	for _, v := range cases {
		rv := reflect.TypeOf(v.value)
		b.Run(v.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = isExportedType(rv)
			}
		})
	}
}
