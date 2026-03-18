package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/modfile"

	migration "github.com/cosmos/cosmos-sdk/tools/migrate"
)

const cosmosSDKModulePath = "github.com/cosmos/cosmos-sdk"
const bridgeSDKVersion = "v0.53.6"

var semverMajorMinorPattern = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.`)

var staleReplaceModules = []string{
	"github.com/cosmos/cosmos-sdk",
	"cosmossdk.io/client/v2",
	"cosmossdk.io/core",
	"cosmossdk.io/store",
	"github.com/cometbft/cometbft",
	"github.com/cosmos/iavl",
	"github.com/prometheus/client_golang",
	"github.com/prometheus/common",
	"cosmossdk.io/x/evidence",
	"cosmossdk.io/x/upgrade",
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
		log.Debug().Msgf("checking dir %q", dir)
		if _, err := os.Stat(dir); err != nil {
			panic(fmt.Sprintf("error with directory: %q: %v\n", dir, err))
		}
	}

	log.Debug().Msgf("starting v50+ -> v54 migration in dir %q", dir)
	if err := prepareTargetSDKVersion(dir); err != nil {
		panic(fmt.Sprintf("error preparing migration target: %v\n", err))
	}

	args := migration.MigrateArgs{
		// --- go.mod operations ---
		GoModRemoval:           removals,
		GoModAddition:          additions,
		GoModReplacements:      replacements,
		GoModUpdates:           moduleUpdates,
		StripLocalPathReplaces: true, // Remove monorepo-relative replaces (../, ./, etc.)

		// --- AST: import rewrites ---
		ImportUpdates:  importReplacements,
		ImportWarnings: importWarnings,

		// --- AST: type/struct changes ---
		TypeUpdates:        typeReplacements,
		FieldRemovals:      fieldRemovals,
		FieldModifications: fieldModifications,

		// --- AST: function arg changes ---
		ArgUpdates:     callUpdates,
		ArgSurgeries:   argSurgeries,
		CallArgEdits:   callArgEdits,
		ComplexUpdates: complexUpdates,

		// --- AST: statement/block removal ---
		StatementRemovals: statementRemovals,
		MapEntryRemovals:  mapEntryRemovals,

		// --- Text-level replacements (post-AST) ---
		TextReplacements: textReplacements,

		// --- File operations ---
		FileRemovals: fileRemovals,
	}
	if err := migration.Migrate(dir, args); err != nil {
		panic(fmt.Sprintf("error migrating: %v\n", err))
	}

	log.Info().Msg("migration complete — run `goimports -w . && go mod tidy` to finalize")
}

func prepareTargetSDKVersion(dir string) error {
	goModPath := filepath.Join(dir, "go.mod")
	_, version, err := parsedModuleVersion(goModPath, cosmosSDKModulePath)
	if err != nil {
		return err
	}

	major, minor, err := parseMajorMinor(version)
	if err != nil {
		return fmt.Errorf("unsupported %s version %q: %w", cosmosSDKModulePath, version, err)
	}

	if major == 0 && minor == 53 {
		if err := dropStaleReplaceModules(goModPath, staleReplaceModules...); err != nil {
			return err
		}
		return nil
	}

	if major == 0 && minor >= 50 && minor < 53 {
		log.Info().Msgf(
			"bridging %s from %s to %s before applying v54 migration",
			cosmosSDKModulePath,
			version,
			bridgeSDKVersion,
		)
		if err := setRequiredModuleVersion(goModPath, cosmosSDKModulePath, bridgeSDKVersion); err != nil {
			return fmt.Errorf("bridge %s to %s: %w", cosmosSDKModulePath, bridgeSDKVersion, err)
		}
		if err := dropStaleReplaceModules(goModPath, staleReplaceModules...); err != nil {
			return err
		}
		return nil
	}

	if major == 0 && minor < 50 {
		return fmt.Errorf(
			"target requires %s %s; this tool supports v0.50.x through v0.53.x inputs",
			cosmosSDKModulePath,
			version,
		)
	}

	return fmt.Errorf(
		"target requires %s %s; this tool supports v0.50.x through v0.53.x inputs and should not be run on already-migrated targets",
		cosmosSDKModulePath,
		version,
	)
}

func parsedModuleVersion(goModPath, modulePath string) (*modfile.File, string, error) {
	goMod, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", goModPath, err)
	}

	mod, err := modfile.Parse(goModPath, goMod, nil)
	if err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", goModPath, err)
	}

	for _, req := range mod.Require {
		if req.Mod.Path == modulePath {
			return mod, req.Mod.Version, nil
		}
	}

	return nil, "", fmt.Errorf("%s is not required in %s", modulePath, goModPath)
}

func setRequiredModuleVersion(goModPath, modulePath, version string) error {
	goMod, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", goModPath, err)
	}

	mod, err := modfile.Parse(goModPath, goMod, nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", goModPath, err)
	}

	if err := mod.AddRequire(modulePath, version); err != nil {
		return fmt.Errorf("set %s=%s: %w", modulePath, version, err)
	}

	formatted, err := mod.Format()
	if err != nil {
		return fmt.Errorf("format %s: %w", goModPath, err)
	}

	if err := os.WriteFile(goModPath, formatted, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", goModPath, err)
	}

	return nil
}

func dropStaleReplaceModules(goModPath string, modulePaths ...string) error {
	goMod, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", goModPath, err)
	}

	mod, err := modfile.Parse(goModPath, goMod, nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", goModPath, err)
	}

	modified := false
	for _, rep := range mod.Replace {
		for _, modulePath := range modulePaths {
			if rep.Old.Path != modulePath {
				continue
			}
			if err := mod.DropReplace(rep.Old.Path, rep.Old.Version); err != nil {
				return fmt.Errorf("drop replace for %s in %s: %w", modulePath, goModPath, err)
			}
			modified = true
			break
		}
	}
	if !modified {
		return nil
	}

	formatted, err := mod.Format()
	if err != nil {
		return fmt.Errorf("format %s: %w", goModPath, err)
	}
	if err := os.WriteFile(goModPath, formatted, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", goModPath, err)
	}

	return nil
}

func parseMajorMinor(version string) (int, int, error) {
	matches := semverMajorMinorPattern.FindStringSubmatch(version)
	if matches == nil {
		return 0, 0, fmt.Errorf("expected semantic version, got %q", version)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse major version: %w", err)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, fmt.Errorf("parse minor version: %w", err)
	}

	return major, minor, nil
}
