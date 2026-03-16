package migration

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/ast/astutil"
)

type GoModReplacement struct {
	Module      string
	Replacement string
	Version     string
}

type (
	// GoModUpdate defines a mapping of module path to the version it should be updated to.
	GoModUpdate map[string]string
	// GoModAddition is a mapping of module name to version string.
	GoModAddition map[string]string
	GoModRemoval  []string
)

// Warning represents a migration issue that cannot be automated and requires manual attention.
type Warning struct {
	// File is the file path where the warning was triggered.
	File string
	// Message is the human-readable warning message.
	Message string
}

// ImportWarning defines an import path pattern that should trigger a warning
// and optionally remove the import from the AST.
type ImportWarning struct {
	// ImportPrefix is the import path prefix to detect.
	ImportPrefix string
	// Message is the warning message to display when the import is found.
	Message string
	// AlsoRemove, if true, removes the matching import from the AST after warning.
	// This is useful for imports that cannot be auto-rewritten but should be stripped
	// to allow the code to compile (e.g., x/group under commercial license).
	AlsoRemove bool
}

type MigrateArgs struct {
	// --- go.mod operations ---
	GoModRemoval           GoModRemoval
	GoModAddition          GoModAddition
	GoModReplacements      []GoModReplacement
	GoModUpdates           GoModUpdate
	StripLocalPathReplaces bool // If true, drop all replace directives with local-path targets (../, ./, etc.)

	// --- AST: import rewrites ---
	ImportUpdates  []ImportReplacement
	ImportWarnings []ImportWarning

	// --- AST: type/struct changes ---
	TypeUpdates        []TypeReplacement
	FieldRemovals      []StructFieldRemoval
	FieldModifications []StructFieldModification

	// --- AST: function arg changes ---
	ArgUpdates        []FunctionArgUpdate
	ArgSurgeries      []ArgSurgeryWithAST
	PlainArgSurgeries []ArgSurgery
	CallArgEdits      []CallArgRemoval
	ComplexUpdates    []ComplexFunctionReplacement

	// --- AST: statement/block removal ---
	StatementRemovals []StatementRemoval
	MapEntryRemovals  []MapEntryRemoval

	// --- Text-level replacements (post-AST) ---
	TextReplacements []TextReplacement

	// --- File operations ---
	FileRemovals []FileRemoval
}

// Migrate migrates all the code in the specified directory.
func Migrate(directory string, args MigrateArgs) error {
	// find all Go files in the directory
	var goFiles []string
	var goModuleFiles []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == "go.mod" {
			goModuleFiles = append(goModuleFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Phase 1: File removals (before AST processing so we don't process dead files)
	if err := applyFileRemovals(directory, args.FileRemovals); err != nil {
		return fmt.Errorf("error removing files: %w", err)
	}

	// Rebuild file list after removals
	goFiles = nil
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var warnings []Warning
	var warningsMu sync.Mutex

	// Phase 2: AST transformations
	if err := updateFiles(goFiles, args, &warnings, &warningsMu); err != nil {
		return fmt.Errorf("error updating files: %w", err)
	}

	// Phase 3: Text-level replacements (for patterns too complex for AST but reliable as text)
	if len(args.TextReplacements) > 0 {
		for _, filePath := range goFiles {
			if _, err := applyTextReplacements(filePath, args.TextReplacements); err != nil {
				return fmt.Errorf("error applying text replacements to %s: %w", filePath, err)
			}
		}
	}

	// Phase 4: go.mod updates
	if err := updateGoModules(goModuleFiles, args.GoModUpdates, args.GoModRemoval, args.GoModReplacements, args.GoModAddition, args.StripLocalPathReplaces); err != nil {
		return fmt.Errorf("error updating go.mod files: %w", err)
	}

	// Print warnings
	if len(warnings) > 0 {
		log.Warn().Msg("=== WARNINGS ===")
		for _, w := range warnings {
			log.Warn().Msgf("  [%s] %s", w.File, w.Message)
		}
	}

	return nil
}

func updateGoModules(goModFiles []string, updates GoModUpdate, removals GoModRemoval, replacements []GoModReplacement, additions GoModAddition, stripLocalReplaces bool) error {
	eg := errgroup.Group{}
	for _, filePath := range goModFiles {
		eg.Go(func() error {
			log.Debug().Msgf("processing %s", filePath)
			file, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			modFile, err := modfile.Parse(filePath, file, nil)
			if err != nil {
				return fmt.Errorf("error parsing %s: %w", filePath, err)
			}
			modified := false

			// Strip local-path replace directives (../, ./, etc.)
			// These exist because the module normally lives inside a monorepo.
			// When migrating a standalone copy, they must be removed so that
			// go mod tidy resolves modules from the registry.
			if stripLocalReplaces && modFile.Replace != nil {
				for _, rep := range modFile.Replace {
					newPath := rep.New.Path
					if strings.HasPrefix(newPath, "../") || strings.HasPrefix(newPath, "./") || newPath == ".." || newPath == "." {
						log.Debug().Msgf("stripping local replace: %s => %s", rep.Old.Path, newPath)
						if err := modFile.DropReplace(rep.Old.Path, rep.Old.Version); err != nil {
							return fmt.Errorf("error dropping replace for %s: %w", rep.Old.Path, err)
						}
						modified = true
					}
				}
			}

			for mod, ver := range additions {
				if err := modFile.AddRequire(mod, ver); err != nil {
					return fmt.Errorf("error adding %s requirement: %w", mod, err)
				}
				modified = true
			}
			for _, module := range modFile.Require {
				if slices.Contains(removals, module.Mod.Path) {
					if err := modFile.DropRequire(module.Mod.Path); err != nil {
						return fmt.Errorf("error removing %s: %w", module.Mod.Path, err)
					}
					modified = true
				}
				if newVersion, ok := updates[module.Mod.Path]; ok {
					if err := modFile.AddRequire(module.Mod.Path, newVersion); err != nil {
						return fmt.Errorf("error updating %s: %w", module.Mod.Path, err)
					}
					modified = true
				}
			}
			for _, replacement := range replacements {
				if err := modFile.AddReplace(replacement.Module, "", replacement.Replacement, replacement.Version); err != nil {
					return fmt.Errorf("error adding replace for %s: %w", replacement.Module, err)
				}
				modified = true
			}
			if modified {
				bz, err := modFile.Format()
				if err != nil {
					return fmt.Errorf("error formatting modified go.mod: %w", err)
				}
				err = os.WriteFile(filePath, bz, 0o600)
				if err != nil {
					return fmt.Errorf("error writing modified go.mod: %w", err)
				}
			}
			return nil
		})
	}
	return eg.Wait()
}

// checkImportWarnings scans a file's imports for patterns that require attention.
// If AlsoRemove is set on a warning, the matching import is removed from the AST.
// Returns warnings and whether any imports were removed (changed).
func checkImportWarnings(filePath string, fset *token.FileSet, node *ast.File, importWarnings []ImportWarning) ([]Warning, bool) {
	type importToRemove struct {
		name string // alias/blank ("_") or "" for unaliased
		path string
	}
	var warnings []Warning
	var toRemove []importToRemove
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		for _, iw := range importWarnings {
			if strings.HasPrefix(importPath, iw.ImportPrefix) {
				warnings = append(warnings, Warning{
					File:    filePath,
					Message: fmt.Sprintf("import %q detected — %s", importPath, iw.Message),
				})
				if iw.AlsoRemove {
					name := ""
					if imp.Name != nil {
						name = imp.Name.Name
					}
					toRemove = append(toRemove, importToRemove{name: name, path: importPath})
				}
			}
		}
	}
	changed := false
	for _, ir := range toRemove {
		if astutil.DeleteNamedImport(fset, node, ir.name, ir.path) {
			changed = true
			log.Debug().Msgf("removed import %s %q (AlsoRemove)", ir.name, ir.path)
		} else {
			log.Warn().Msgf("failed to remove import %s %q — not found in AST", ir.name, ir.path)
		}
	}
	if changed {
		ast.SortImports(fset, node)
	}
	return warnings, changed
}

func updateFiles(goFiles []string, args MigrateArgs, warnings *[]Warning, warningsMu *sync.Mutex) error {
	eg := errgroup.Group{}
	for _, filePath := range goFiles {
		eg.Go(func() error {
			log.Debug().Msgf("processing %s", filePath)
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("error parsing %s: %w", filePath, err)
			}

			changed := false

			// Check for import warnings (and optionally remove matching imports)
			if len(args.ImportWarnings) > 0 {
				fileWarnings, c := checkImportWarnings(filePath, fset, node, args.ImportWarnings)
				if len(fileWarnings) > 0 {
					warningsMu.Lock()
					*warnings = append(*warnings, fileWarnings...)
					warningsMu.Unlock()
				}
				changed = changed || c
			}

			// Import rewrites
			if c, err := updateImports(node, args.ImportUpdates); err != nil {
				return fmt.Errorf("error updating imports in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Type renames
			if c, err := updateStructs(node, args.TypeUpdates); err != nil {
				return fmt.Errorf("error updating structs in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Struct field removals
			if c, err := updateStructFieldRemovals(node, args.FieldRemovals); err != nil {
				return fmt.Errorf("error removing struct fields in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Struct field modifications
			if c, err := updateStructFieldModifications(node, args.FieldModifications); err != nil {
				return fmt.Errorf("error modifying struct fields in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Simple arg truncation
			if c, err := updateFunctionCalls(node, args.ArgUpdates); err != nil {
				return fmt.Errorf("error updating function calls in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Arg surgery (positional remove/insert/wrap) — AST callback variant
			if c, err := updateArgSurgeryAST(node, args.ArgSurgeries); err != nil {
				return fmt.Errorf("error applying arg surgery in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Arg surgery — declarative variant (with $ARG{N} placeholders)
			if c, err := updateArgSurgery(node, args.PlainArgSurgeries); err != nil {
				return fmt.Errorf("error applying plain arg surgery in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Call arg edits (remove/add args from specific calls)
			if c, err := updateCallArgRemovals(node, args.CallArgEdits); err != nil {
				return fmt.Errorf("error editing call args in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Complex function replacements
			if c, err := updateComplexFunctions(fset, node, args.ComplexUpdates); err != nil {
				return fmt.Errorf("error updating complex functions in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Statement removals
			if c, err := updateStatementRemovals(node, args.StatementRemovals); err != nil {
				return fmt.Errorf("error removing statements in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			// Map entry removals
			if c, err := updateMapEntryRemovals(node, args.MapEntryRemovals); err != nil {
				return fmt.Errorf("error removing map entries in %s: %w", filePath, err)
			} else {
				changed = changed || c
			}

			if changed {
				buf := new(bytes.Buffer)
				err = format.Node(buf, fset, node)
				if err != nil {
					return fmt.Errorf("error formatting modified code: %w", err)
				}
				err = os.WriteFile(filePath, buf.Bytes(), 0o600)
				if err != nil {
					return fmt.Errorf("error writing modified code: %w", err)
				}
			}
			return nil
		})
	}
	return eg.Wait()
}
