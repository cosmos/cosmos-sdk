package migrationtool

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

var sdkVersionRE = regexp.MustCompile(`github\.com/cosmos/cosmos-sdk\s+((?:v|V)[^\s]+)`)

type RepoScan struct {
	Root          string
	Files         []RepoFile
	GoFiles       []RepoFile
	GoModFiles    []RepoFile
	SDKVersions   []SDKVersion
	SelectedSpecs []SpecMatch
	Blockers      []SpecMatch
}

type RepoFile struct {
	RelPath string
	Content string
}

type SDKVersion struct {
	GoModPath string
	Version   string
	Supported bool
}

type SpecMatch struct {
	Spec           Spec
	ImportMatches  []Match
	PatternMatches []Match
}

type Match struct {
	Needle string
	File   string
}

func ScanRepo(repoRoot string, specDir string) (*RepoScan, error) {
	specs, err := LoadSpecs(specDir)
	if err != nil {
		return nil, err
	}

	files, err := readRepoFiles(repoRoot)
	if err != nil {
		return nil, err
	}

	scan := &RepoScan{
		Root:       repoRoot,
		Files:      files,
		GoFiles:    filterFiles(files, func(path string) bool { return strings.HasSuffix(path, ".go") }),
		GoModFiles: filterFiles(files, func(path string) bool { return filepath.Base(path) == "go.mod" }),
	}
	scan.SDKVersions = detectSDKVersions(scan.GoModFiles)

	for _, spec := range specs {
		match := detectSpec(scan, spec)
		if !match.applies() {
			continue
		}
		if filepath.Base(spec.SourcePath) == "group.yaml" {
			scan.Blockers = append(scan.Blockers, match)
			continue
		}
		scan.SelectedSpecs = append(scan.SelectedSpecs, match)
	}

	return scan, nil
}

func RunScan(w io.Writer, repoRoot string, specDir string) error {
	scan, err := ScanRepo(repoRoot, specDir)
	if err != nil {
		return err
	}

	renderHeader(w, "Scan")
	fmt.Fprintf(w, "Repo: %s\n\n", scan.Root)
	writeVersionSection(w, scan.SDKVersions)
	writeSelectionSection(w, scan.Blockers, scan.SelectedSpecs)
	return nil
}

func RunPlan(w io.Writer, repoRoot string, specDir string) error {
	scan, err := ScanRepo(repoRoot, specDir)
	if err != nil {
		return err
	}

	renderHeader(w, "Plan")
	fmt.Fprintf(w, "Repo: %s\n\n", scan.Root)
	writeVersionSection(w, scan.SDKVersions)
	writeSelectionSection(w, scan.Blockers, scan.SelectedSpecs)
	if len(scan.Blockers) > 0 {
		fmt.Fprintln(w, "Plan halted because blocking specs were detected.")
		return nil
	}

	if len(scan.SelectedSpecs) == 0 {
		fmt.Fprintln(w, "No migration specs were selected.")
		return nil
	}

	fmt.Fprintln(w, "Execution order:")
	for index, match := range scan.SelectedSpecs {
		spec := match.Spec
		fmt.Fprintf(w, "%d. %s (%s)\n", index+1, spec.Title, filepath.Base(spec.SourcePath))
		fmt.Fprintf(w, "   - id: %s\n", spec.ID)
		fmt.Fprintf(w, "   - summary: %s\n", compact(spec.Description))
		fmt.Fprintf(w, "   - planned changes: %s\n", describeChanges(spec.Changes))
		if len(spec.ManualSteps) > 0 {
			fmt.Fprintf(w, "   - manual steps: %d\n", len(spec.ManualSteps))
		}
		fmt.Fprintf(w, "   - verification checks: %d\n", len(spec.Verification.MustNotImport)+len(spec.Verification.MustNotContain)+len(spec.Verification.MustContain))
	}

	return nil
}

func detectSpec(scan *RepoScan, spec Spec) SpecMatch {
	match := SpecMatch{Spec: spec}

	for _, needle := range spec.Detection.Imports {
		if hit, ok := firstContentMatch(scan.GoFiles, needle); ok {
			match.ImportMatches = append(match.ImportMatches, Match{Needle: needle, File: hit.RelPath})
		}
	}

	for _, needle := range spec.Detection.Patterns {
		if hit, ok := firstPatternMatch(scan.Files, needle); ok {
			match.PatternMatches = append(match.PatternMatches, Match{Needle: needle, File: hit.RelPath})
		}
	}

	return match
}

func (m SpecMatch) applies() bool {
	return len(m.ImportMatches) > 0 || len(m.PatternMatches) > 0
}

func readRepoFiles(root string) ([]RepoFile, error) {
	var files []RepoFile

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}

		if !isTextCandidate(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		files = append(files, RepoFile{
			RelPath: filepath.ToSlash(relPath),
			Content: string(content),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})

	return files, nil
}

func isTextCandidate(path string) bool {
	base := filepath.Base(path)
	return base == "go.mod" || filepath.Ext(path) == ".go"
}

func filterFiles(files []RepoFile, keep func(path string) bool) []RepoFile {
	var filtered []RepoFile
	for _, file := range files {
		if keep(file.RelPath) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func detectSDKVersions(goMods []RepoFile) []SDKVersion {
	var versions []SDKVersion

	for _, file := range goMods {
		matches := sdkVersionRE.FindStringSubmatch(file.Content)
		if len(matches) == 0 {
			continue
		}
		version := matches[1]
		versions = append(versions, SDKVersion{
			GoModPath: file.RelPath,
			Version:   version,
			Supported: isSupportedSDKVersion(version),
		})
	}

	return versions
}

func isSupportedSDKVersion(version string) bool {
	prefixes := []string{"v0.50.", "v0.51.", "v0.52.", "v0.53.", "v0.54."}
	return slices.ContainsFunc(prefixes, func(prefix string) bool {
		return strings.HasPrefix(version, prefix)
	})
}

func firstContentMatch(files []RepoFile, needle string) (RepoFile, bool) {
	for _, file := range files {
		if strings.Contains(file.Content, needle) {
			return file, true
		}
	}
	return RepoFile{}, false
}

func firstPatternMatch(files []RepoFile, needle string) (RepoFile, bool) {
	for _, file := range files {
		if strings.Contains(file.Content, needle) || strings.Contains(file.RelPath, needle) {
			return file, true
		}
	}
	return RepoFile{}, false
}

func writeVersionSection(w io.Writer, versions []SDKVersion) {
	fmt.Fprintln(w, "SDK versions:")
	if len(versions) == 0 {
		fmt.Fprintln(w, "- no cosmos-sdk requirement found in scanned go.mod files")
		fmt.Fprintln(w)
		return
	}

	for _, version := range versions {
		status := "unsupported"
		if version.Supported {
			status = "supported"
		}
		fmt.Fprintf(w, "- %s: %s (%s)\n", version.GoModPath, version.Version, status)
	}
	fmt.Fprintln(w)
}

func writeSelectionSection(w io.Writer, blockers []SpecMatch, selected []SpecMatch) {
	fmt.Fprintln(w, "Detected specs:")

	if len(blockers) == 0 && len(selected) == 0 {
		fmt.Fprintln(w, "- none")
		fmt.Fprintln(w)
		return
	}

	for _, match := range blockers {
		fmt.Fprintf(w, "- BLOCKER %s (%s)\n", match.Spec.Title, filepath.Base(match.Spec.SourcePath))
		writeMatchReasons(w, match)
	}
	for _, match := range selected {
		fmt.Fprintf(w, "- %s (%s)\n", match.Spec.Title, filepath.Base(match.Spec.SourcePath))
		writeMatchReasons(w, match)
	}
	fmt.Fprintln(w)
}

func writeMatchReasons(w io.Writer, match SpecMatch) {
	for _, hit := range match.ImportMatches {
		fmt.Fprintf(w, "  import match: %q in %s\n", hit.Needle, hit.File)
	}
	for _, hit := range match.PatternMatches {
		fmt.Fprintf(w, "  pattern match: %q in %s\n", hit.Needle, hit.File)
	}
}

func describeChanges(changes Changes) string {
	var parts []string

	if len(changes.GoMod.Update) > 0 {
		parts = append(parts, fmt.Sprintf("%d go.mod updates", len(changes.GoMod.Update)))
	}
	if len(changes.GoMod.Remove) > 0 {
		parts = append(parts, fmt.Sprintf("%d go.mod removals", len(changes.GoMod.Remove)))
	}
	if changes.GoMod.StripLocalReplaces {
		parts = append(parts, "strip local replaces")
	}
	if len(changes.Imports.Rewrites) > 0 {
		parts = append(parts, fmt.Sprintf("%d import rewrites", len(changes.Imports.Rewrites)))
	}
	if len(changes.Imports.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("%d import warnings", len(changes.Imports.Warnings)))
	}
	if len(changes.FileRemovals) > 0 {
		parts = append(parts, fmt.Sprintf("%d file removals", len(changes.FileRemovals)))
	}
	if len(changes.StatementRemovals) > 0 {
		parts = append(parts, fmt.Sprintf("%d statement removals", len(changes.StatementRemovals)))
	}
	if len(changes.MapEntryRemovals) > 0 {
		parts = append(parts, fmt.Sprintf("%d map entry removals", len(changes.MapEntryRemovals)))
	}
	if len(changes.CallArgEdits) > 0 {
		parts = append(parts, fmt.Sprintf("%d call argument edits", len(changes.CallArgEdits)))
	}
	if len(changes.TextReplacements) > 0 {
		parts = append(parts, fmt.Sprintf("%d text replacements", len(changes.TextReplacements)))
	}

	if len(parts) == 0 {
		return "manual review only"
	}
	return strings.Join(parts, ", ")
}

func renderHeader(w io.Writer, title string) {
	fmt.Fprintf(w, "%s\n%s\n\n", title, strings.Repeat("=", len(title)))
}

func compact(input string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
}
