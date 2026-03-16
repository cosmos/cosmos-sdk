package migration

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/rs/zerolog/log"
)

// StatementRemoval defines a pattern for removing assignment statements or expression
// statements that match a specific receiver + function call pattern.
//
// For example, to remove `app.CircuitKeeper = circuitkeeper.NewKeeper(...)`:
//
//	StatementRemoval{
//	    AssignTarget: "app.CircuitKeeper",
//	}
//
// Or to remove `app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)`:
//
//	StatementRemoval{
//	    CallPattern: "app.BaseApp.SetCircuitBreaker",
//	}
//
// Both multi-line statements (with block comments between them) are handled.
type StatementRemoval struct {
	// AssignTarget matches assignment statements where the LHS is this dotted path.
	// e.g., "app.CircuitKeeper" matches `app.CircuitKeeper = ...`
	AssignTarget string
	// CallPattern matches expression statements where the call is this dotted path.
	// e.g., "app.BaseApp.SetCircuitBreaker" matches `app.BaseApp.SetCircuitBreaker(...)`
	CallPattern string
	// IncludeFollowing defines how many statements after the matched one should also be removed.
	// Use this for closely associated statements like:
	//   app.CircuitKeeper = circuitkeeper.NewKeeper(...) // matched
	//   app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper) // IncludeFollowing: 1
	IncludeFollowing int
	// IncludePrecedingComment removes the comment group immediately before the statement.
	IncludePrecedingComment bool
	// IncludePrecedingAssign matches and removes a preceding variable assignment.
	// e.g., "groupConfig" will remove `groupConfig := group.DefaultConfig()` before the target.
	IncludePrecedingAssign string
	// IncludePrecedingBlock removes preceding block/multi-line comments.
	IncludePrecedingBlock bool
}

// updateStatementRemovals walks block statements and removes matching statements.
func updateStatementRemovals(node *ast.File, removals []StatementRemoval) (bool, error) {
	if len(removals) == 0 {
		return false, nil
	}

	modified := false

	ast.Inspect(node, func(n ast.Node) bool {
		blockStmt, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		var newList []ast.Stmt
		skipUntil := -1

		for i, stmt := range blockStmt.List {
			if i <= skipUntil {
				continue
			}

			shouldRemove := false
			var matchedRemoval *StatementRemoval

			for j := range removals {
				removal := &removals[j]
				if matchesRemoval(stmt, removal) {
					shouldRemove = true
					matchedRemoval = removal
					break
				}
			}

			if shouldRemove {
				log.Debug().Msgf("Removing statement matching removal rule")
				modified = true

				// Also remove IncludeFollowing statements
				if matchedRemoval.IncludeFollowing > 0 {
					skipUntil = i + matchedRemoval.IncludeFollowing
				}

				// Remove preceding variable assignment if specified
				if matchedRemoval.IncludePrecedingAssign != "" && len(newList) > 0 {
					last := newList[len(newList)-1]
					if matchesPrecedingAssign(last, matchedRemoval.IncludePrecedingAssign) {
						newList = newList[:len(newList)-1]
					}
				}

				continue
			}

			newList = append(newList, stmt)
		}

		if len(newList) != len(blockStmt.List) {
			blockStmt.List = newList
		}

		return true
	})

	return modified, nil
}

// matchesRemoval checks if a statement matches a removal pattern.
func matchesRemoval(stmt ast.Stmt, removal *StatementRemoval) bool {
	if removal.AssignTarget != "" {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			if len(assignStmt.Lhs) > 0 {
				lhs := exprToString(assignStmt.Lhs[0])
				if lhs == removal.AssignTarget {
					return true
				}
			}
		}
	}

	if removal.CallPattern != "" {
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
			if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
				callStr := exprToString(callExpr.Fun)
				if callStr == removal.CallPattern {
					return true
				}
			}
		}
	}

	return false
}

// matchesPrecedingAssign checks if a statement is a short variable declaration
// for the given variable name, e.g., `groupConfig := group.DefaultConfig()`.
func matchesPrecedingAssign(stmt ast.Stmt, varName string) bool {
	if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
		if assignStmt.Tok == token.DEFINE && len(assignStmt.Lhs) > 0 {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
				return ident.Name == varName
			}
		}
	}
	return false
}

// CallArgRemoval defines removal of specific arguments from a function call.
// Used to remove entries from variadic calls like storetypes.NewKVStoreKeys(),
// module.NewManager(), SetOrderEndBlockers(), etc.
type CallArgRemoval struct {
	// FuncPattern is a dotted call pattern to match, e.g., "storetypes.NewKVStoreKeys".
	// For method calls, this can also be a method name like "SetOrderEndBlockers".
	FuncPattern string
	// MethodName matches the method being called (for receiver-style calls like app.ModuleManager.SetOrderEndBlockers).
	MethodName string
	// ArgsToRemove lists the Go expression strings of arguments to remove.
	// Each entry is matched against the string representation of each argument.
	ArgsToRemove []string
	// ArgsToAdd lists Go expression strings to add at specific positions.
	ArgsToAdd []ArgAddition
}

// ArgAddition defines an argument to add at a specific position in a call.
type ArgAddition struct {
	// Position is the 0-indexed position to insert at. Use -1 to append.
	Position int
	// Expr is the Go expression string.
	Expr string
}

// updateCallArgRemovals finds function calls and removes/adds specific arguments.
func updateCallArgRemovals(node *ast.File, removals []CallArgRemoval) (bool, error) {
	if len(removals) == 0 {
		return false, nil
	}

	modified := false

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		for _, removal := range removals {
			if !matchesCallPattern(callExpr, removal) {
				continue
			}

			// Remove matching arguments
			var newArgs []ast.Expr
			for _, arg := range callExpr.Args {
				argStr := exprToString(arg)
				shouldRemove := false
				for _, toRemove := range removal.ArgsToRemove {
					if argStr == toRemove {
						shouldRemove = true
						log.Debug().Msgf("Removing arg %q from call", argStr)
						break
					}
				}
				if !shouldRemove {
					newArgs = append(newArgs, arg)
				}
			}

			if len(newArgs) != len(callExpr.Args) {
				modified = true
			}

			// Add new arguments
			for _, addition := range removal.ArgsToAdd {
				expr, err := parseExprSafe(addition.Expr)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to parse arg addition: %s", addition.Expr)
					continue
				}

				// Copy position info from existing args so the printer
				// doesn't split the new SelectorExpr across lines.
				if len(newArgs) > 0 {
					setExprPos(expr, newArgs[0].Pos())
				} else {
					setExprPos(expr, callExpr.Lparen+1)
				}

				if addition.Position == -1 || addition.Position >= len(newArgs) {
					newArgs = append(newArgs, expr)
				} else {
					// Insert at position
					newArgs = append(newArgs[:addition.Position+1], newArgs[addition.Position:]...)
					newArgs[addition.Position] = expr
				}
				modified = true
				log.Debug().Msgf("Added arg %q to call", addition.Expr)
			}

			callExpr.Args = newArgs
			break
		}

		return true
	})

	return modified, nil
}

// matchesCallPattern checks if a call expression matches the specified pattern.
func matchesCallPattern(callExpr *ast.CallExpr, removal CallArgRemoval) bool {
	if removal.FuncPattern != "" {
		callStr := exprToString(callExpr.Fun)
		if callStr == removal.FuncPattern {
			return true
		}
	}

	if removal.MethodName != "" {
		// Match any receiver with this method name
		if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == removal.MethodName {
				return true
			}
		}
	}

	return false
}

// MapEntryRemoval defines removal of entries from a map literal.
type MapEntryRemoval struct {
	// MapVarName matches the variable name being assigned (e.g., "maccPerms").
	MapVarName string
	// KeysToRemove lists the Go expression strings of keys to remove.
	KeysToRemove []string
}

// updateMapEntryRemovals finds map literals and removes matching key-value pairs.
func updateMapEntryRemovals(node *ast.File, removals []MapEntryRemoval) (bool, error) {
	if len(removals) == 0 {
		return false, nil
	}

	modified := false

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for value specs (var declarations) and assign statements
		switch v := n.(type) {
		case *ast.ValueSpec:
			for _, removal := range removals {
				for _, name := range v.Names {
					if name.Name == removal.MapVarName {
						for _, value := range v.Values {
							if compLit, ok := value.(*ast.CompositeLit); ok {
								if removeMapEntries(compLit, removal.KeysToRemove) {
									modified = true
								}
							}
						}
					}
				}
			}
		case *ast.AssignStmt:
			for _, removal := range removals {
				for _, lhs := range v.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && ident.Name == removal.MapVarName {
						for _, rhs := range v.Rhs {
							if compLit, ok := rhs.(*ast.CompositeLit); ok {
								if removeMapEntries(compLit, removal.KeysToRemove) {
									modified = true
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	return modified, nil
}

// removeMapEntries removes key-value pairs from a composite literal.
func removeMapEntries(compLit *ast.CompositeLit, keysToRemove []string) bool {
	modified := false
	var newElts []ast.Expr

	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			newElts = append(newElts, elt)
			continue
		}

		keyStr := exprToString(kv.Key)
		shouldRemove := false
		for _, toRemove := range keysToRemove {
			if keyStr == toRemove {
				shouldRemove = true
				log.Debug().Msgf("Removing map entry with key %q", keyStr)
				break
			}
		}

		if !shouldRemove {
			newElts = append(newElts, elt)
		} else {
			modified = true
		}
	}

	if modified {
		compLit.Elts = newElts
	}
	return modified
}

// exprToString converts an AST expression to its approximate string representation.
// Used for pattern matching — not for code generation.
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.CallExpr:
		return exprToString(e.Fun) + "(...)"
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.UnaryExpr:
		return e.Op.String() + exprToString(e.X)
	case *ast.IndexExpr:
		return exprToString(e.X) + "[" + exprToString(e.Index) + "]"
	case *ast.BasicLit:
		return e.Value
	case *ast.CompositeLit:
		if e.Type != nil {
			return exprToString(e.Type) + "{...}"
		}
		return "{...}"
	default:
		return ""
	}
}

// parseExprSafe parses a Go expression string into an AST expression.
func parseExprSafe(expr string) (ast.Expr, error) {
	// Wrap in a var declaration so the parser can handle arbitrary expressions
	src := "package p\nvar _ = " + expr
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return nil, err
	}

	// Extract the expression from the parsed file
	if len(f.Decls) > 0 {
		if genDecl, ok := f.Decls[0].(*ast.GenDecl); ok && len(genDecl.Specs) > 0 {
			if valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec); ok && len(valueSpec.Values) > 0 {
				return valueSpec.Values[0], nil
			}
		}
	}

	return nil, fmt.Errorf("failed to extract expression from parsed source: %s", expr)
}

// setExprPos recursively sets position info on an AST expression so that
// the Go printer keeps it on one line instead of splitting across lines.
func setExprPos(expr ast.Expr, pos token.Pos) {
	switch e := expr.(type) {
	case *ast.Ident:
		e.NamePos = pos
	case *ast.SelectorExpr:
		setExprPos(e.X, pos)
		e.Sel.NamePos = pos
	case *ast.StarExpr:
		e.Star = pos
		setExprPos(e.X, pos)
	case *ast.CallExpr:
		setExprPos(e.Fun, pos)
		e.Lparen = pos
		e.Rparen = pos
	case *ast.BasicLit:
		e.ValuePos = pos
	}
}
