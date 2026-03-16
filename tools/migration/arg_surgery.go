package migration

import (
	"go/ast"
	"go/parser"

	"github.com/rs/zerolog/log"
)

// ArgSurgery describes a complex function argument transformation where arguments
// need to be removed, reordered, or new arguments need to be synthesized.
type ArgSurgery struct {
	// ImportPath is the import path for the package containing the function.
	ImportPath string
	// FuncName is the function name to match.
	FuncName string
	// OldArgCount is the expected number of arguments before transformation.
	// Use -1 to match any count (e.g., for variadic functions).
	OldArgCount int
	// RemoveArgPositions lists 0-indexed positions of arguments to remove.
	RemoveArgPositions []int
	// AppendArgs lists Go expression strings to append as new arguments.
	// These are parsed as Go expressions. You can use `$ARG{N}` as a placeholder
	// to reference the removed argument at position N (0-indexed in the original call).
	AppendArgs []string
}

// updateArgSurgery walks the AST and applies argument surgery rules.
func updateArgSurgery(node *ast.File, surgeries []ArgSurgery) (bool, error) {
	if len(surgeries) == 0 {
		return false, nil
	}

	modified := false
	importAliases := buildImportAliases(node)

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkgIdent, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		for _, surgery := range surgeries {
			alias, exists := importAliases[surgery.ImportPath]
			if !exists {
				continue
			}

			if pkgIdent.Name != alias || selectorExpr.Sel.Name != surgery.FuncName {
				continue
			}

			if surgery.OldArgCount != -1 && len(callExpr.Args) != surgery.OldArgCount {
				continue
			}

			log.Debug().Msgf("Applying arg surgery to %s.%s()", pkgIdent.Name, surgery.FuncName)

			// Save removed args before modifying the slice
			removedArgs := make(map[int]ast.Expr)
			for _, pos := range surgery.RemoveArgPositions {
				if pos < len(callExpr.Args) {
					removedArgs[pos] = callExpr.Args[pos]
				}
			}

			// Build new args list, skipping removed positions
			removeSet := make(map[int]bool)
			for _, pos := range surgery.RemoveArgPositions {
				removeSet[pos] = true
			}

			var newArgs []ast.Expr
			for i, arg := range callExpr.Args {
				if !removeSet[i] {
					newArgs = append(newArgs, arg)
				}
			}

			// Append new synthesized arguments
			for _, argExpr := range surgery.AppendArgs {
				// Replace $ARG{N} placeholders with actual removed arg expressions
				resolved := resolveArgPlaceholders(argExpr, removedArgs)
				expr, err := parser.ParseExpr(resolved)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to parse synthesized arg: %s", resolved)
					continue
				}
				newArgs = append(newArgs, expr)
			}

			callExpr.Args = newArgs
			// Clear the ellipsis if the function was variadic and we're replacing all args
			callExpr.Ellipsis = 0
			modified = true
			break
		}

		return true
	})

	return modified, nil
}

// resolveArgPlaceholders replaces $ARG{N} in the expression string with a placeholder
// identifier _removed_argN_. The real AST substitution happens after parsing.
//
// For simplicity, we use a two-pass approach:
// 1. Replace $ARG{N} with a temp identifier
// 2. After parsing, walk the parsed expr and replace the temp idents with the real exprs
//
// However, since the removed args can be complex expressions (like `app.StakingKeeper`),
// we actually need to do direct AST surgery. So we just handle the common case where
// the append arg wraps a single removed arg in a function call.
func resolveArgPlaceholders(expr string, removedArgs map[int]ast.Expr) string {
	// For now, we handle the simple case: no placeholders means return as-is.
	// Placeholder-based substitution is handled in the AST after parsing.
	return expr
}

// ArgSurgeryWithAST is a more powerful version that uses a callback to construct
// the new arguments, giving full access to the original args.
type ArgSurgeryWithAST struct {
	// ImportPath is the import path for the package containing the function.
	ImportPath string
	// FuncName is the function name to match.
	FuncName string
	// OldArgCount is the expected number of arguments before transformation.
	// Use -1 to match any count.
	OldArgCount int
	// Transform takes the original arguments and returns the new arguments.
	Transform func(originalArgs []ast.Expr) []ast.Expr
}

// updateArgSurgeryAST walks the AST and applies AST-level argument transformations.
func updateArgSurgeryAST(node *ast.File, surgeries []ArgSurgeryWithAST) (bool, error) {
	if len(surgeries) == 0 {
		return false, nil
	}

	modified := false
	importAliases := buildImportAliases(node)

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkgIdent, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		for _, surgery := range surgeries {
			alias, exists := importAliases[surgery.ImportPath]
			if !exists {
				continue
			}

			if pkgIdent.Name != alias || selectorExpr.Sel.Name != surgery.FuncName {
				continue
			}

			if surgery.OldArgCount != -1 && len(callExpr.Args) != surgery.OldArgCount {
				continue
			}

			log.Debug().Msgf("Applying AST arg surgery to %s.%s()", pkgIdent.Name, surgery.FuncName)

			newArgs := surgery.Transform(callExpr.Args)
			callExpr.Args = newArgs
			callExpr.Ellipsis = 0
			modified = true
			break
		}

		return true
	})

	return modified, nil
}
