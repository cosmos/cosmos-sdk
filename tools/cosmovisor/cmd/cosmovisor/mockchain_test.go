package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/tools/cosmovisor/internal"
)

type MockChainSetup struct {
	Genesis        string
	GovUpgrades    map[string]string
	ManualUpgrades map[string]string // to be added with the add-upgrade command
	Config         *cosmovisor.Config
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
	// create gov upgrade wrappers
	for name, args := range m.GovUpgrades {
		upgradeDir := filepath.Join(cosmovisorDir, "upgrades", name, "bin")
		require.NoError(t, os.MkdirAll(upgradeDir, 0o755))
		require.NoError(t,
			os.WriteFile(filepath.Join(upgradeDir, "mockd"),
				[]byte(mockNodeWrapper(args)), 0o755),
		)
	}
	// create manual upgrade wrappers
	manualUpgradeDir := filepath.Join(dir, "manual-upgrades")
	require.NoError(t, os.MkdirAll(manualUpgradeDir, 0o755))
	for name, args := range m.ManualUpgrades {
		filename := filepath.Join(manualUpgradeDir, name)
		require.NoError(t, os.WriteFile(filename, []byte(mockNodeWrapper(args)), 0o755))
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
		Genesis: "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":30}'",
		GovUpgrades: map[string]string{
			"gov1": "--block-time 1s --upgrade-plan '{\"name\":\"gov2\",\"height\":40}'",
		},
		ManualUpgrades: map[string]string{
			"manual10": "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":30}'",
			"manual20": "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":30}'",
		},
		Config: &cosmovisor.Config{
			PollInterval:        time.Second,
			RestartAfterUpgrade: true,
		},
	}.Setup(t)

	var callbackQueue []func()
	testCallback := func() {
		for _, cb := range callbackQueue {
			cb()
		}
		callbackQueue = nil // reset for next test
	}
	callbackQueue = append(callbackQueue, func() {
		go func() {
			time.Sleep(2 * time.Second) // wait for startup
			rootCmd := NewRootCmd()
			rootCmd.SetArgs([]string{
				"add-upgrade",
				"manual20",
				filepath.Join(mockchainDir, "manual-upgrades", "manual20"),
				"--upgrade-height",
				"20",
				"--cosmovisor-config",
				cfgFile,
			})
			rootCmd.SetOut(os.Stdout)
			rootCmd.SetErr(os.Stderr)
			require.NoError(t, rootCmd.Execute())
		}()
	})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"run", "--home", mockchainDir, "--cosmovisor-config", cfgFile})
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetErr(os.Stderr)
		ctx := internal.WithTestCallback(context.Background(), testCallback)
		require.NoError(t, rootCmd.ExecuteContext(ctx))
		wg.Done()
	}()
	wg.Wait()

	// TODO:
	// - [x] add callback on restart for checking state
	// - [ ] add manual upgrade (manual20) at height 20
	// - [ ] then add other manual upgrades manual10 at height 10 and manual30 at height 30 as a batch
	// - [ ] manual20 should get picked up and the process should restart with halt-height 20
	// - [ ] then manual10 should get picked up and the process should restart with halt-height 10
	// - [ ] when manual10 gets applied, it should restart with halt-height 20
	// - [ ] when manual20 gets applied, it should restart with no halt-height
	// - [ ] and then manual20 should trigger gov2 upgrade at height 40
	// - [ ]
}
