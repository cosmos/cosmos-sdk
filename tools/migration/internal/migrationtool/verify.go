package migrationtool

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

type VerifyOptions struct {
	RunGoModTidy bool
	RunGoBuild   bool
	RunGoTest    bool
}

type VerificationFailure struct {
	SpecID  string
	File    string
	Pattern string
	Reason  string
}

func RunVerify(w io.Writer, repoRoot string, specDir string, opts VerifyOptions) error {
	scan, err := ScanRepo(repoRoot, specDir)
	if err != nil {
		return err
	}

	renderHeader(w, "Verify")
	fmt.Fprintf(w, "Repo: %s\n\n", scan.Root)
	writeVersionSection(w, scan.SDKVersions)
	writeSelectionSection(w, scan.Blockers, scan.SelectedSpecs)

	if len(scan.Blockers) > 0 {
		fmt.Fprintln(w, "Verification halted because blocking specs were detected.")
		return fmt.Errorf("blocking spec detected")
	}

	failures := VerifyStaticChecks(scan)
	if len(failures) > 0 {
		fmt.Fprintln(w, "Static verification failures:")
		for _, failure := range failures {
			fmt.Fprintf(w, "- [%s] %s in %s (%s)\n", failure.SpecID, failure.Pattern, failure.File, failure.Reason)
		}
	} else {
		fmt.Fprintln(w, "Static verification passed.")
	}
	fmt.Fprintln(w)

	if opts.RunGoModTidy {
		if err := runRepoCommand(w, repoRoot, "go", "mod", "tidy"); err != nil {
			return err
		}
	}
	if opts.RunGoBuild {
		if err := runRepoCommand(w, repoRoot, "go", "build", "./..."); err != nil {
			return err
		}
	}
	if opts.RunGoTest {
		if err := runRepoCommand(w, repoRoot, "go", "test", "./..."); err != nil {
			return err
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("static verification failed")
	}

	return nil
}

func VerifyStaticChecks(scan *RepoScan) []VerificationFailure {
	var failures []VerificationFailure

	for _, match := range scan.SelectedSpecs {
		spec := match.Spec
		for _, needle := range spec.Verification.MustNotImport {
			if hit, ok := firstContentMatch(scan.GoFiles, needle); ok {
				failures = append(failures, VerificationFailure{
					SpecID:  spec.ID,
					File:    hit.RelPath,
					Pattern: needle,
					Reason:  "unexpected import still present",
				})
			}
		}
		for _, check := range spec.Verification.MustNotContain {
			if hit, ok := firstPatternMatch(scan.Files, check.Pattern); ok {
				failures = append(failures, VerificationFailure{
					SpecID:  spec.ID,
					File:    hit.RelPath,
					Pattern: check.Pattern,
					Reason:  "forbidden pattern still present",
				})
			}
		}
		for _, check := range spec.Verification.MustContain {
			if _, ok := firstPatternMatch(scan.Files, check.Pattern); !ok {
				failures = append(failures, VerificationFailure{
					SpecID:  spec.ID,
					File:    filepath.ToSlash(scan.Root),
					Pattern: check.Pattern,
					Reason:  "required pattern not found",
				})
			}
		}
	}

	return failures
}

func runRepoCommand(w io.Writer, repoRoot string, name string, args ...string) error {
	fmt.Fprintf(w, "Running: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = repoRoot

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(w, "%s\n", output.String())
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}

	fmt.Fprintf(w, "%s\n", output.String())
	return nil
}
