package migration

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"golang.org/x/sync/errgroup"
)

type MigrateArgs struct {
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
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
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
	if err := updateFiles(goFiles, args); err != nil {
		return err
	}
	return nil
}

func updateFiles(goFiles []string, args MigrateArgs) error {
	eg := errgroup.Group{}
	for _, filePath := range goFiles {
		eg.Go(func() error {
			log.Debug().Msgf("processing %s", filePath)
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("error parsing %s: %v", filePath, err)
			}

			importsChanged, err := updateImports(node, args.ImportUpdates)
			if err != nil {
				return fmt.Errorf("error updating imports in %s: %v", filePath, err)
			}
			structsChanged, err := updateStructs(node, args.TypeUpdates)
			if err != nil {
				return fmt.Errorf("error updating structs in %s: %v", filePath, err)
			}
			callsChanged, err := updateFunctionCalls(node, args.ArgUpdates)
			if err != nil {
				return fmt.Errorf("error updating function calls in %s: %v", filePath, err)
			}
			complexCallsChanged, err := updateComplexFunctions(fset, node, args.ComplexUpdates)
			if err != nil {
				return fmt.Errorf("error updating complex function calls in %s: %v", filePath, err)
			}

			changed := importsChanged || structsChanged || callsChanged || complexCallsChanged
			if changed {
				buf := new(bytes.Buffer)
				err = format.Node(buf, fset, node)
				if err != nil {
					return fmt.Errorf("error formatting modified code: %v", err)
				}
				err = os.WriteFile(filePath, buf.Bytes(), 0644)
				if err != nil {
					return fmt.Errorf("error writing modified code: %v", err)
				}
			}
			return nil
		})
	}
	return eg.Wait()
}
