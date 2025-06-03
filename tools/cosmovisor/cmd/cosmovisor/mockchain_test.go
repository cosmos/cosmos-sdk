package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/tools/cosmovisor"
)

type MockChainSetup struct {
	Genesis  string
	Upgrades map[string]string
	Config   *cosmovisor.Config
}

func mockNodeWrapper(args string) string {
	return fmt.Sprintf(
		`#!/usr/bin/env bash
set -e

echo "$@"
exec mock_node %s "$@" 
`, args)
}

func (m MockChainSetup) Setup(t *testing.T) (string, string) {
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
	mockdPath := filepath.Join(genDir, "mockd")
	require.NoError(t,
		os.WriteFile(mockdPath, []byte(mockNodeWrapper(m.Genesis)), 0o755),
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

	// update config and save it
	if m.Config == nil {
		m.Config = &cosmovisor.Config{}
	}
	m.Config.Name = "mockd"
	m.Config.Home = dir
	m.Config.DataBackupPath = dir
	cfgFile, err := m.Config.Export()
	require.NoError(t, err)
	t.Logf("Cosmovisor config: %s", cfgFile)

	return dir, cfgFile
}

func TestMockChain(t *testing.T) {
	mockchainDir, cfgFile := MockChainSetup{
		Genesis: "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":15}'",
		Upgrades: map[string]string{
			"gov1":    "--halt-height 20 --block-time 1s --upgrade-plan '{\"name\":\"gov2\",\"height\":30}'",
			"manual1": "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":15}'",
		},
		Config: &cosmovisor.Config{
			PollInterval: time.Second,
		},
	}.Setup(t)

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"run", "--home", mockchainDir, "--cosmovisor-config", cfgFile})
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	require.NoError(t, rootCmd.ExecuteContext(context.Background()))
}
