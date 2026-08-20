package systemtests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRestoreOriginalConfigs covers restoreOriginalConfigs directly: every node config
// file edited after setup is put back from its backup. The SetupChain/ResetChain wiring
// around it needs a node binary and is exercised by the system test suites.
func TestRestoreOriginalConfigs(t *testing.T) {
	const (
		outputDir  = "testnet"
		project    = "simd"
		nodesCount = 2
	)

	oldWorkDir := WorkDir
	t.Cleanup(func() { WorkDir = oldWorkDir })
	WorkDir = t.TempDir()

	sut := &SystemUnderTest{
		outputDir:         outputDir,
		projectName:       project,
		initialNodesCount: nodesCount,
		nodesCount:        nodesCount,
	}

	// setup: pristine config plus the backup SetupChain takes
	for i := 0; i < nodesCount; i++ {
		configDir := filepath.Join(WorkDir, sut.nodePath(i), "config")
		require.NoError(t, os.MkdirAll(configDir, 0o750))
		for _, tomlFile := range nodeConfigFiles {
			path := filepath.Join(configDir, tomlFile)
			require.NoError(t, os.WriteFile(path, []byte("pruning = \"default\"\n"), 0o600))
			MustCopyFile(path, path+".orig")
		}
	}

	// a test edits the config, as configureBlockPruning does
	for i := 0; i < nodesCount; i++ {
		for _, tomlFile := range nodeConfigFiles {
			path := filepath.Join(WorkDir, sut.nodePath(i), "config", tomlFile)
			require.NoError(t, os.WriteFile(path, []byte("pruning = \"everything\"\n"), 0o600))
		}
	}

	restoreOriginalConfigs(t, sut)

	for i := 0; i < nodesCount; i++ {
		for _, tomlFile := range nodeConfigFiles {
			got, err := os.ReadFile(filepath.Join(WorkDir, sut.nodePath(i), "config", tomlFile))
			require.NoError(t, err)
			require.Equal(t, "pruning = \"default\"\n", string(got), "node%d %s not restored", i, tomlFile)
		}
	}
}
