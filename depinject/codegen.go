package depinject

import (
	"fmt"
	"go/ast"
)

func (c *container) createIdent(namePrefix string) *ast.Ident {
	return c.doCreateIdent(namePrefix, nil)
}

func (c *container) doCreateIdent(namePrefix string, handle interface{}) *ast.Ident {
	// TODO reserved names: keywords, builtin types, imports
	v := ast.NewIdent(namePrefix)
	i := 2
	for {
		_, ok := c.idents[v]
		if !ok {
			c.idents[v] = handle
			if handle != nil {
				c.reverseIdents[handle] = v
			}
			return v
		}

		v = ast.NewIdent(fmt.Sprintf("%s%d", namePrefix, i))
		i++
	}
}

func (c *container) getOrCreateIdent(namePrefix string, handle interface{}) (v *ast.Ident, created bool) {
	if v, ok := c.reverseIdents[handle]; ok {
		return v, false
	}

	return c.doCreateIdent(namePrefix, handle), true
}

func (c *container) codegenStmt(stmt ast.Stmt) {
	c.codegenBody.List = append(c.codegenBody.List, stmt)
}
