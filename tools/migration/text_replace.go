package migration

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// TextReplacement defines a simple text find-and-replace operation.
// This is used as a last resort for patterns that are too complex for AST manipulation
// but have reliable textual patterns (e.g., `app.BaseApp.GRPCQueryRouter()` -> `app.GRPCQueryRouter()`).
type TextReplacement struct {
	// Old is the text to find.
	Old string
	// New is the replacement text.
	New string
	// FileMatch restricts this replacement to files whose base name matches.
	// If empty, the replacement applies to all files.
	FileMatch string
	// RequiresContains lists substrings that must be present in the file content
	// before this replacement is applied. This keeps narrow fixture-oriented
	// rewrites from firing on custom chain code with superficially similar text.
	RequiresContains []string
}

// FileRemoval defines a file to delete during migration.
type FileRemoval struct {
	// FileName is the base name of the file to delete (e.g., "ante.go").
	FileName string
	// ContainsMustMatch is a string that must be present in the file for it to be deleted.
	// This prevents accidental deletion of files with the same name but different content.
	ContainsMustMatch string
}

// applyTextReplacements applies text-level find-and-replace to a file's content.
// Returns true if any replacements were made.
func applyTextReplacements(filePath string, replacements []TextReplacement) (bool, error) {
	if len(replacements) == 0 {
		return false, nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	original := string(content)
	result := original

	for _, r := range replacements {
		if r.FileMatch != "" && !strings.HasSuffix(filepath.ToSlash(filePath), "/"+r.FileMatch) {
			continue
		}
		if len(r.RequiresContains) > 0 {
			missingRequiredContent := false
			for _, required := range r.RequiresContains {
				if !strings.Contains(result, required) {
					missingRequiredContent = true
					break
				}
			}
			if missingRequiredContent {
				continue
			}
		}
		result = strings.ReplaceAll(result, r.Old, r.New)
	}

	if result != original {
		err = os.WriteFile(filePath, []byte(result), 0o600)
		if err != nil {
			return false, err
		}
		log.Debug().Msgf("Applied text replacements to %s", filePath)
		return true, nil
	}

	return false, nil
}

// applyFileRemovals deletes files matching the removal rules.
func applyFileRemovals(directory string, removals []FileRemoval) error {
	for _, removal := range removals {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || info.Name() != removal.FileName {
				return nil
			}

			// If a must-match pattern is specified, check file content
			if removal.ContainsMustMatch != "" {
				content, err := os.ReadFile(path) //nolint:gosec // path is safe; migration runs on trusted local project files
				if err != nil {
					return nil // skip files we can't read
				}
				if !strings.Contains(string(content), removal.ContainsMustMatch) {
					return nil // file doesn't match the safety check
				}
			}

			log.Info().Msgf("Removing file: %s", path)
			return os.Remove(path) //nolint:gosec // path is safe; migration runs on trusted local project files
		})
		if err != nil {
			return err
		}
	}
	return nil
}
