package main

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

func main() {
	directory := "."
	args := os.Args
	if len(args) > 1 {
		directory = args[1]
	}

	// find all Go files in the directory
	var goFiles []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Msg("error walking the directory")
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("error walking the directory")
		os.Exit(1)
	}
	if err := updateFiles(goFiles); err != nil {
		log.Error().Err(err).Msg("error updating files")
		os.Exit(1)
	}
}

func updateFiles(goFiles []string) error {
	eg := errgroup.Group{}
	for _, filePath := range goFiles {
		eg.Go(func() error {
			log.Debug().Msgf("processing %s", filePath)
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("error parsing %s: %v", filePath, err)
			}

			importsChanged, err := updateImports(node, importReplacements)
			if err != nil {
				return fmt.Errorf("error updating imports in %s: %v", filePath, err)
			}
			structsChanged, err := updateStructs(node, typeReplacements)
			if err != nil {
				return fmt.Errorf("error updating structs in %s: %v", filePath, err)
			}
			callsChanged, err := updateFunctionCalls(node, functionUpdates)
			if err != nil {
				return fmt.Errorf("error updating function calls in %s: %v", filePath, err)
			}
			complexCallsChanged, err := updateComplexFunctions(fset, node, complexReplacements)
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
