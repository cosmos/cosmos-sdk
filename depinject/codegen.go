package depinject

import (
	"go/ast"
)

func (c *container) codegenStmt(stmt ast.Stmt) {
	c.funcGen.Func.Body.List = append(c.funcGen.Func.Body.List, stmt)
}
