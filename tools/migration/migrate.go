package migration

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
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

type MigrateArgs struct {
	GoModRemoval      GoModRemoval
	GoModAddition     GoModAddition
	GoModReplacements []GoModReplacement
	// GoModUpdates defines the list of modules to update.
	GoModUpdates GoModUpdate
	// ArgUpdates defines the necessary changes where a function has reduced its arguments.
	ArgUpdates []FunctionArgUpdate
	// ComplexUpdates defines the rules for replacing function calls with custom replacement logic.
	ComplexUpdates []ComplexFunctionReplacement
	// ImportUpdates defines the list of import replacement rules to update old import paths to new ones.
	ImportUpdates []ImportReplacement
	// TypeUpdates updates type names.
	TypeUpdates []TypeReplacement
}

// Migrate migrates the all the code in the specified directory.
func Migrate(directory string, args MigrateArgs) error {
	// find all Go files in the directory
	var goFiles []string
	var goModuleFiles []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		if !info.IsDir() && strings.HasSuffix(path, "go.mod") {
			goModuleFiles = append(goModuleFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := updateFiles(goFiles, args); err != nil {
		return fmt.Errorf("error updating files: %w", err)
	}
	if err := updateGoModules(goModuleFiles, args.GoModUpdates, args.GoModRemoval, args.GoModReplacements, args.GoModAddition); err != nil {
		return fmt.Errorf("error updating go.mod files: %w", err)
	}
	return nil
}

func updateGoModules(goModFiles []string, updates GoModUpdate, removals GoModRemoval, replacements []GoModReplacement, additions GoModAddition) error {
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
			for mod, ver := range additions {
				if err := modFile.AddRequire(mod, ver); err != nil {
					return fmt.Errorf("error adding %s requirement: %w", mod, err)
				}
				modified = true
			}
			// loop through all the modules in the go.mod file.
			// we don't care about indirect modules, we only want to update direct dependencies.
			for _, module := range modFile.Require {
				if module.Indirect {
					continue
				}
				if slices.Contains(removals, module.Mod.Path) {
					if err := modFile.DropRequire(module.Mod.Path); err != nil {
						return fmt.Errorf("error removing %s: %w", module.Mod.Path, err)
					}
					modified = true
				}
				// if this module is one we want to update it, we call AddRequire, which updates the version.
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
			// if we modified the go mod: format, write, tidy.
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

func updateFiles(goFiles []string, args MigrateArgs) error {
	eg := errgroup.Group{}
	for _, filePath := range goFiles {
		eg.Go(func() error {
			log.Debug().Msgf("processing %s", filePath)
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("error parsing %s: %w", filePath, err)
			}

			structsChanged, err := updateStructs(node, args.TypeUpdates)
			if err != nil {
				return fmt.Errorf("error updating structs in %s: %w", filePath, err)
			}
			callsChanged, err := updateFunctionCalls(node, args.ArgUpdates)
			if err != nil {
				return fmt.Errorf("error updating function calls in %s: %w", filePath, err)
			}
			complexCallsChanged, err := updateComplexFunctions(fset, node, args.ComplexUpdates)
			if err != nil {
				return fmt.Errorf("error updating complex function calls in %s: %w", filePath, err)
			}
			importsChanged, err := updateImports(node, args.ImportUpdates)
			if err != nil {
				return fmt.Errorf("error updating imports in %s: %w", filePath, err)
			}

			changed := importsChanged || structsChanged || callsChanged || complexCallsChanged
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
