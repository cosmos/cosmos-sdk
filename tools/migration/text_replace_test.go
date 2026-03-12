package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyTextReplacements(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		replacements []TextReplacement
		wantModified bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name:    "simple replacement",
			content: "app.BaseApp.GRPCQueryRouter()",
			replacements: []TextReplacement{
				{Old: "app.BaseApp.GRPCQueryRouter()", New: "app.GRPCQueryRouter()"},
			},
			wantModified: true,
			wantContains: []string{"app.GRPCQueryRouter()"},
			wantMissing:  []string{"app.BaseApp.GRPCQueryRouter()"},
		},
		{
			name:    "multiple replacements",
			content: "app.BaseApp.GRPCQueryRouter() and app.BaseApp.Simulate",
			replacements: []TextReplacement{
				{Old: "app.BaseApp.GRPCQueryRouter()", New: "app.GRPCQueryRouter()"},
				{Old: "app.BaseApp.Simulate", New: "app.Simulate"},
			},
			wantModified: true,
			wantContains: []string{"app.GRPCQueryRouter()", "app.Simulate"},
			wantMissing:  []string{"app.BaseApp.GRPCQueryRouter()", "app.BaseApp.Simulate"},
		},
		{
			name:    "deletion replacement (empty new string)",
			content: "line1\n\t\tCircuitKeeper: &app.CircuitKeeper,\nline3",
			replacements: []TextReplacement{
				{Old: "\t\tCircuitKeeper: &app.CircuitKeeper,\n", New: ""},
			},
			wantModified: true,
			wantContains: []string{"line1\nline3"},
			wantMissing:  []string{"CircuitKeeper"},
		},
		{
			name:    "no match — no modification",
			content: "unchanged content",
			replacements: []TextReplacement{
				{Old: "not present", New: "replacement"},
			},
			wantModified: false,
			wantContains: []string{"unchanged content"},
		},
		{
			name:         "empty replacements",
			content:      "some content",
			replacements: []TextReplacement{},
			wantModified: false,
			wantContains: []string{"some content"},
		},
		{
			name: "multi-line replacement (SetModuleVersionMap)",
			content: "\tapp.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())\n\treturn app.ModuleManager.InitGenesis(",
			replacements: []TextReplacement{
				{
					Old: "\tapp.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())\n\treturn app.ModuleManager.InitGenesis(",
					New: "\terr := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn app.ModuleManager.InitGenesis(",
				},
			},
			wantModified: true,
			wantContains: []string{"err :=", "if err != nil", "return nil, err"},
		},
		{
			name:    "FileMatch applies replacement to matching filename",
			content: "\t\"time\"\nother content",
			replacements: []TextReplacement{
				{Old: "\t\"time\"\n", New: "", FileMatch: "test.go"},
			},
			wantModified: true,
			wantContains: []string{"other content"},
			wantMissing:  []string{"\"time\""},
		},
		{
			name:    "FileMatch skips non-matching filename",
			content: "\t\"time\"\nother content",
			replacements: []TextReplacement{
				{Old: "\t\"time\"\n", New: "", FileMatch: "app_config.go"},
			},
			wantModified: false,
			wantContains: []string{"\"time\"", "other content"},
		},
		{
			name:    "EpochsKeeper init pattern",
			content: "\tapp.EpochsKeeper = epochskeeper.NewKeeper(\n\t\truntime.NewKVStoreService(keys[epochstypes.StoreKey]),\n\t\tappCodec,\n\t)\n\n\tapp.EpochsKeeper.SetHooks(",
			replacements: []TextReplacement{
				{Old: "app.EpochsKeeper = epochskeeper.NewKeeper(", New: "epochsKeeper := epochskeeper.NewKeeper("},
				{Old: "\tapp.EpochsKeeper.SetHooks(", New: "\tapp.EpochsKeeper = &epochsKeeper\n\n\tapp.EpochsKeeper.SetHooks("},
			},
			wantModified: true,
			wantContains: []string{
				"epochsKeeper := epochskeeper.NewKeeper(",
				"app.EpochsKeeper = &epochsKeeper",
				"app.EpochsKeeper.SetHooks(",
			},
			wantMissing: []string{"app.EpochsKeeper = epochskeeper.NewKeeper("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.go")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("write temp file: %v", err)
			}

			modified, err := applyTextReplacements(tmpFile, tt.replacements)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if modified != tt.wantModified {
				t.Errorf("modified = %v, want %v", modified, tt.wantModified)
			}

			result, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("read result: %v", err)
			}
			output := string(result)

			for _, s := range tt.wantContains {
				if !strings.Contains(output, s) {
					t.Errorf("output should contain %q, got:\n%s", s, output)
				}
			}
			for _, s := range tt.wantMissing {
				if strings.Contains(output, s) {
					t.Errorf("output should NOT contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestApplyFileRemovals(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string // filename -> content
		removals       []FileRemoval
		wantRemoved    []string
		wantRemaining  []string
	}{
		{
			name: "remove file matching content check",
			files: map[string]string{
				"ante.go": `package main
import circuitante "cosmossdk.io/x/circuit/ante"
func f() { circuitante.NewCircuitBreakerDecorator() }`,
			},
			removals: []FileRemoval{
				{FileName: "ante.go", ContainsMustMatch: "circuitante"},
			},
			wantRemoved: []string{"ante.go"},
		},
		{
			name: "skip file that does not match content check",
			files: map[string]string{
				"ante.go": `package main
func f() { /* no circuit stuff */ }`,
			},
			removals: []FileRemoval{
				{FileName: "ante.go", ContainsMustMatch: "circuitante"},
			},
			wantRemaining: []string{"ante.go"},
		},
		{
			name: "remove file without content check",
			files: map[string]string{
				"old.go": "package main",
			},
			removals: []FileRemoval{
				{FileName: "old.go"},
			},
			wantRemoved: []string{"old.go"},
		},
		{
			name: "only removes matching filename",
			files: map[string]string{
				"ante.go": `package main; // circuitante`,
				"app.go":  `package main; // circuitante`,
			},
			removals: []FileRemoval{
				{FileName: "ante.go", ContainsMustMatch: "circuitante"},
			},
			wantRemoved:   []string{"ante.go"},
			wantRemaining: []string{"app.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o600); err != nil {
					t.Fatalf("write %s: %v", name, err)
				}
			}

			err := applyFileRemovals(tmpDir, tt.removals)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			for _, name := range tt.wantRemoved {
				if _, err := os.Stat(filepath.Join(tmpDir, name)); !os.IsNotExist(err) {
					t.Errorf("file %q should have been removed", name)
				}
			}
			for _, name := range tt.wantRemaining {
				if _, err := os.Stat(filepath.Join(tmpDir, name)); err != nil {
					t.Errorf("file %q should still exist, got: %v", name, err)
				}
			}
		})
	}
}
