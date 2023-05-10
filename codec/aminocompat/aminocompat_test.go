package aminocompat

import (
	"strings"
	"testing"
)

func TestAllClear(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		wantErr string
	}{
		{"nil", nil, "not valid"},
		{"int", 12, ""},
		{"string", "aminocompat", ""},

		// Unsupported basics.
		{"map", map[int]int{}, "not supported"},
		{"complex64", complex64(10 + 1i), "not supported"},
		{"complex128", complex128(10 + 1i), "not supported"},
		{"float32", float32(10), "not supported"},
		{"float64", float64(10), "not supported"},

		// Supported composites
		{
			"struct with 8th level value",
			&s{
				A: &s2nd{B: &s3rd{A: &s4th{A: s5th{A: s6th{A: s7th{A: &s8th{A: "8th", B: 10}}}}}}},
			},
			"",
		},
		{
			"struct with an unexported field but that's unsupported by amino-json",
			&sWithunexportedunsupported{
				a: map[string]int{"a": 10},
				B: 10,
				C: "st1",
			},
			"",
		},

		// Unsupported composites
		{
			"struct with map",
			&sWithMap{
				A: 10,
				B: "ab",
				C: map[string]int{"a": 10},
			},
			"not supported: map",
		},

		{
			"slice of maps",
			&sWithMapSlice{
				A: [][]map[int]int{
					{map[int]int{1: 10}},
				},
			},
			"not supported: map",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := AllClear(tt.in)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected an error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("could not find\n\t%q\nin\n\t%q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

type s8th struct {
	A string
	B int
}

type s7th struct {
	A *s8th
}

type s6th struct {
	A s7th
}

type s5th struct {
	A s6th
}

type s4th struct {
	A s5th
}
type s3rd struct {
	A *s4th
	B string
}
type s2nd struct {
	B *s3rd
	C int
}
type s struct {
	A *s2nd
	B []string
}

type sWithMap struct {
	A int
	B string
	C map[string]int
}

type sWithunexportedunsupported struct {
	a map[string]int
	B int
	C string
}

type sWithMapSlice struct {
	A [][]map[int]int
}
