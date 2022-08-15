package depinject

import (
	"go/parser"
	"testing"

	"gotest.tools/v3/assert"
)

const src1 = `//go:build depinject
package foo

func CodegenTest1() {
}
`

const src2 = `//go:build depinject
package foo

func CodegenTest1() (int, error) {
}
`

const src3 = `//go:build depinject
package foo

import "cosmossdk.io/depinject"

func CodegenTest1() (x int, err error) {
  err = depinject.BuildDebug(depinject.Codegen())
  return
}
`

func TestCodegen(t *testing.T) {
	cases := []struct {
		src         string
		errContains string
	}{
		{
			src1,
			"expected non-empty output",
		},
		{
			src2,
			"expected exactly 2 statements in function body",
		},
		{
			src3,
			"",
		},
	}

	for _, s := range cases {
		cfg, err := newDebugConfig()
		assert.NilError(t, err)
		cfg.codegenLoc = &location{
			name: "CodegenTest1",
			pkg:  "foo",
			file: "",
			line: 0,
		}
		f, err := parser.ParseFile(cfg.fset, "test1.go", s.src, parser.ParseComments|parser.AllErrors)
		assert.NilError(t, err)
		err = cfg.startCodegen(f)
		if s.errContains != "" {
			assert.ErrorContains(t, err, s.errContains)
		} else {
			assert.NilError(t, err)
		}
	}
}
