package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	"gotest.tools/v3/assert"
)

func TestReservedIdents(t *testing.T) {
	const src = `
package mypkg

func Build(a, a2 int, b string) (c int, err error) {}
`
	file, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	assert.NilError(t, err)

	fileGen, err := NewFileGen(file, "example.com/mypkg")
	assert.NilError(t, err)
	gen := fileGen.PatchFuncDecl("Build")
	assert.Assert(t, gen != nil)

	// go keywords get a suffix
	assert.Equal(t, "type2", gen.CreateIdent("type").Name)
	assert.Equal(t, "package2", gen.CreateIdent("package").Name)
	assert.Equal(t, "goto2", gen.CreateIdent("goto").Name)

	// param and result names get suffixes
	assert.Equal(t, "a3", gen.CreateIdent("a").Name)
	assert.Equal(t, "b2", gen.CreateIdent("b").Name)
	assert.Equal(t, "c2", gen.CreateIdent("c").Name)
	assert.Equal(t, "err2", gen.CreateIdent("err").Name)
}
