package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/tools/migration/internal/migrationtool"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("migrate-to-v54", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	repo := fs.String("repo", "", "path to the target chain repository")
	specDir := fs.String("spec-dir", defaultSpecDir(), "path to the v50-to-v54 spec directory")
	runGoModTidy := fs.Bool("go-mod-tidy", false, "run go mod tidy during verify")
	runGoBuild := fs.Bool("go-build", false, "run go build ./... during verify")
	runGoTest := fs.Bool("go-test", false, "run go test ./... during verify")

	if err := fs.Parse(args); err != nil {
		return err
	}

	command := "plan"
	if remaining := fs.Args(); len(remaining) > 0 {
		command = remaining[0]
	}

	if *repo == "" {
		return fmt.Errorf("--repo is required")
	}

	absRepo, err := filepath.Abs(*repo)
	if err != nil {
		return err
	}

	absSpecDir, err := filepath.Abs(*specDir)
	if err != nil {
		return err
	}

	switch command {
	case "scan":
		return migrationtool.RunScan(os.Stdout, absRepo, absSpecDir)
	case "plan":
		return migrationtool.RunPlan(os.Stdout, absRepo, absSpecDir)
	case "verify":
		opts := migrationtool.VerifyOptions{
			RunGoModTidy: *runGoModTidy,
			RunGoBuild:   *runGoBuild,
			RunGoTest:    *runGoTest,
		}
		return migrationtool.RunVerify(os.Stdout, absRepo, absSpecDir, opts)
	default:
		return fmt.Errorf("unknown command %q, expected scan, plan, or verify", command)
	}
}

func defaultSpecDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "tools/migration/migration-spec/v50-to-v54"
	}

	candidates := []string{
		filepath.Join(wd, "migration-spec", "v50-to-v54"),
		filepath.Join(wd, "tools", "migration", "migration-spec", "v50-to-v54"),
	}

	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && stat.IsDir() {
			return candidate
		}
	}

	return candidates[len(candidates)-1]
}
