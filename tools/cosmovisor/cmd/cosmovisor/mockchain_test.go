package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockChainSetup struct {
	Genesis  string
	Upgrades map[string]string
}

func mockNodeWrapper(args string) string {
	return fmt.Sprintf(`
#!/usr/bin/env bash
set -e

exec mock_node %s "$@"
`, args)
}

func (m MockChainSetup) Setup(t *testing.T) {
	dir, err := os.MkdirTemp("", "mockchain")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})
	// create data directory
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "data"), 0o755))
	cosmovisorDir := filepath.Join(dir, "cosmovisor")
	// create genesis wrapper
	genDir := filepath.Join(cosmovisorDir, "genesis", "bin")
	require.NoError(t, os.MkdirAll(genDir, 0o755))
	require.NoError(t,
		os.WriteFile(filepath.Join(genDir, "mockd"),
			[]byte(mockNodeWrapper(m.Genesis)), 0o755),
	)
	// create upgrade wrappers
	for name, args := range m.Upgrades {
		upgradeDir := filepath.Join(cosmovisorDir, "upgrades", name, "bin")
		require.NoError(t, os.MkdirAll(upgradeDir, 0o755))
		require.NoError(t,
			os.WriteFile(filepath.Join(upgradeDir, "mockd"),
				[]byte(mockNodeWrapper(args)), 0o755),
		)
	}
}

func TestMockChain(t *testing.T) {
	MockChainSetup{
		Genesis: "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":14}'",
		Upgrades: map[string]string{
			"gov1":    "--halt-height 20 --block-time 1s --upgrade-plan '{\"name\":\"gov2\",\"height\":30}'",
			"manual1": "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":15}'",
		},
	}.Setup(t)
	//dir, err := os.Getwd()
	//require.NoError(t, err)
	//if !strings.HasSuffix(dir, "tools/cosmovisor/cmd/cosmovisor") {
	//	t.Fatalf("expected to be in tools/cosmovisor/cmd/cosmovisor, got %s", dir)
	//}
	//// switch to the root of the cosmovisor project
	//t.Chdir(filepath.Join(dir, "..", ".."))
	//// clean up previous test runs
	//require.NoError(t, exec.Command("make", "clean").Run())
}
