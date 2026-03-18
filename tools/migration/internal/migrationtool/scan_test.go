package migrationtool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanRepoSelectsSpecsInOrder(t *testing.T) {
	repo := t.TempDir()
	specDir := filepath.Join("..", "..", "migration-spec", "v50-to-v54")

	writeFixtureFile(t, filepath.Join(repo, "go.mod"), `module example.com/chain

go 1.25.8

require github.com/cosmos/cosmos-sdk v0.53.4
`)
	writeFixtureFile(t, filepath.Join(repo, "app", "app.go"), `package app

import (
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

func wire() {
	_ = crisis.ModuleName
	app.CrisisKeeper = nil
	keeper.NewKeeper(
		appCodec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distrKeeper,
		router,
		govConfig,
		authority,
	)
}
`)

	scan, err := ScanRepo(repo, specDir)
	if err != nil {
		t.Fatalf("ScanRepo: %v", err)
	}

	if len(scan.SDKVersions) != 1 {
		t.Fatalf("expected 1 SDK version, got %d", len(scan.SDKVersions))
	}
	if got := scan.SDKVersions[0].Version; got != "v0.53.4" {
		t.Fatalf("unexpected version %q", got)
	}

	var ids []string
	for _, match := range scan.SelectedSpecs {
		ids = append(ids, match.Spec.ID)
	}

	want := []string{"core-sdk-migration", "crisis-removal", "gov-keeper-migration"}
	if strings.Join(ids, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected spec order: got %v want %v", ids, want)
	}
}

func TestVerifyStaticChecksFindsMissingAndForbiddenPatterns(t *testing.T) {
	scan := &RepoScan{
		Root: "/tmp/example",
		Files: []RepoFile{
			{RelPath: "app/app.go", Content: `package app
import "github.com/cosmos/cosmos-sdk/x/gov/keeper"
func wire() { keeper.NewKeeper() }
`},
		},
		GoFiles: []RepoFile{
			{RelPath: "app/app.go", Content: `package app
import "github.com/cosmos/cosmos-sdk/x/gov/keeper"
func wire() { keeper.NewKeeper() }
`},
		},
		SelectedSpecs: []SpecMatch{
			{
				Spec: Spec{
					ID: "gov-keeper-migration",
					Verification: Verification{
						MustNotImport: []string{"github.com/cosmos/cosmos-sdk/x/gov/keeper"},
						MustContain: PatternChecks{
							{Pattern: "NewDefaultCalculateVoteResultsAndVotingPower"},
						},
					},
				},
			},
		},
	}

	failures := VerifyStaticChecks(scan)
	if len(failures) != 2 {
		t.Fatalf("expected 2 failures, got %d", len(failures))
	}
}

func writeFixtureFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
