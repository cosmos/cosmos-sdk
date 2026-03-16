package migration

import (
	"fmt"
	"go/ast"
	"go/parser"
	"regexp"
	"strconv"

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
				// Swap the temp identifiers back to the real removed-arg AST nodes
				expr = replaceArgPlaceholderIdents(expr, removedArgs)
				newArgs = append(newArgs, expr)
			}

			callExpr.Args = newArgs
			// Only clear the ellipsis if the original call had one and the last arg changed.
			// If a surgery preserves a ...args spread, the ellipsis must be kept.
			if callExpr.Ellipsis != 0 {
				origLast := originalLastArg(surgery, callExpr)
				if origLast == nil || len(newArgs) == 0 || newArgs[len(newArgs)-1] != origLast {
					callExpr.Ellipsis = 0
				}
			}
			modified = true
			break
		}

		return true
	})

	return modified, nil
}

// originalLastArg returns the last arg from the original call if it was a spread arg.
// For ArgSurgery, we check whether the last original position was NOT removed.
func originalLastArg(surgery ArgSurgery, callExpr *ast.CallExpr) ast.Expr {
	removeSet := make(map[int]bool)
	for _, pos := range surgery.RemoveArgPositions {
		removeSet[pos] = true
	}
	// The original arg count is known from the surgery; check if the last arg survived.
	lastIdx := surgery.OldArgCount - 1
	if lastIdx >= 0 && !removeSet[lastIdx] {
		// The last original arg wasn't removed — find it in the new arg list.
		// Since we build newArgs by iterating originals in order, the surviving last arg
		// will be at the end (before any AppendArgs).
		surviving := 0
		for i := 0; i < surgery.OldArgCount; i++ {
			if !removeSet[i] {
				surviving++
			}
		}
		if surviving > 0 && surviving <= len(callExpr.Args) {
			return callExpr.Args[surviving-1]
		}
	}
	return nil
}

// argPlaceholderRe matches $ARG{N} tokens in AppendArgs strings.
var argPlaceholderRe = regexp.MustCompile(`\$ARG\{(\d+)\}`)

// resolveArgPlaceholders replaces $ARG{N} tokens in expr with temporary identifiers
// (_removedArg0_, _removedArg1_, etc.) so the string can be parsed as valid Go.
// After parsing, replaceArgPlaceholderIdents must be called on the resulting AST
// to swap those identifiers with the real removed-arg expressions.
func resolveArgPlaceholders(expr string, removedArgs map[int]ast.Expr) string {
	return argPlaceholderRe.ReplaceAllStringFunc(expr, func(match string) string {
		sub := argPlaceholderRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		idx, err := strconv.Atoi(sub[1])
		if err != nil {
			return match
		}
		if _, ok := removedArgs[idx]; !ok {
			log.Warn().Msgf("$ARG{%d} referenced but arg at position %d was not removed", idx, idx)
			return match
		}
		return fmt.Sprintf("_removedArg%d_", idx)
	})
}

// replaceArgPlaceholderIdents walks a parsed AST expression and replaces temporary
// _removedArgN_ identifiers (produced by resolveArgPlaceholders) with the actual
// AST expressions that were removed from the original call.
func replaceArgPlaceholderIdents(expr ast.Expr, removedArgs map[int]ast.Expr) ast.Expr {
	ast.Inspect(expr, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			for i, arg := range node.Args {
				if ident, ok := arg.(*ast.Ident); ok {
					if replaced := matchPlaceholderIdent(ident.Name, removedArgs); replaced != nil {
						node.Args[i] = replaced
					}
				}
			}
		case *ast.CompositeLit:
			for i, elt := range node.Elts {
				if ident, ok := elt.(*ast.Ident); ok {
					if replaced := matchPlaceholderIdent(ident.Name, removedArgs); replaced != nil {
						node.Elts[i] = replaced
					}
				}
				// Also handle key-value expressions inside composite literals
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if ident, ok := kv.Value.(*ast.Ident); ok {
						if replaced := matchPlaceholderIdent(ident.Name, removedArgs); replaced != nil {
							kv.Value = replaced
						}
					}
				}
			}
		}
		return true
	})
	// Top-level expression might itself be a placeholder
	if ident, ok := expr.(*ast.Ident); ok {
		if replaced := matchPlaceholderIdent(ident.Name, removedArgs); replaced != nil {
			return replaced
		}
	}
	return expr
}

// placeholderIdentRe matches the temporary identifiers produced by resolveArgPlaceholders.
var placeholderIdentRe = regexp.MustCompile(`^_removedArg(\d+)_$`)

// matchPlaceholderIdent checks whether name is a _removedArgN_ placeholder and returns
// the corresponding removed-arg AST expression, or nil if it doesn't match.
func matchPlaceholderIdent(name string, removedArgs map[int]ast.Expr) ast.Expr {
	sub := placeholderIdentRe.FindStringSubmatch(name)
	if len(sub) < 2 {
		return nil
	}
	idx, err := strconv.Atoi(sub[1])
	if err != nil {
		return nil
	}
	if expr, ok := removedArgs[idx]; ok {
		return expr
	}
	return nil
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

			origArgs := callExpr.Args
			newArgs := surgery.Transform(origArgs)
			callExpr.Args = newArgs
			// Only clear the ellipsis if the original call had one and the last arg changed.
			if callExpr.Ellipsis != 0 && len(origArgs) > 0 {
				origLast := origArgs[len(origArgs)-1]
				if len(newArgs) == 0 || newArgs[len(newArgs)-1] != origLast {
					callExpr.Ellipsis = 0
				}
			}
			modified = true
			break
		}

		return true
	})

	return modified, nil
}
