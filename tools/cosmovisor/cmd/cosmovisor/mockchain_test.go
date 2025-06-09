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
	Genesis string
	// TODO test setup should be similar to process_test.go
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
	pollInterval := time.Second
	cfg := &cosmovisor.Config{
		PollInterval:        pollInterval,
		RestartAfterUpgrade: true,
		RPCAddress:          "http://localhost:26657", // TODO this should be the default!
	}
	mockchainDir, cfgFile := MockChainSetup{
		Genesis: "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":30}'",
		GovUpgrades: map[string]string{
			"gov1": "--block-time 1s --upgrade-plan '{\"name\":\"gov2\",\"height\":50}'",
		},
		ManualUpgrades: map[string]string{
			"manual10": "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":30}'",
			"manual20": `--block-time 1s --upgrade-plan '{"name":"gov1","height":30}' --block-url "/v1/block" --shutdown-on-upgrade`,
			"manual40": "--block-time 1s --upgrade-plan '{\"name\":\"gov1\",\"height\":50}'",
		},
		Config: cfg,
	}.Setup(t)

	addManualUpgrade1 := func() {
		time.Sleep(pollInterval * 2) // wait for startup
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
	}

	addManualUpgrade2 := func() {
		batchInfo := fmt.Sprintf(`manual10:%s:10,manual40:%s:40`,
			filepath.Join(mockchainDir, "manual-upgrades", "manual10"),
			filepath.Join(mockchainDir, "manual-upgrades", "manual40"),
		)
		time.Sleep(2 * time.Second) // wait for startup
		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{
			"add-batch-upgrade",
			"--upgrade-list",
			batchInfo,
			"--cosmovisor-config",
			cfgFile,
		})
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetErr(os.Stderr)
		require.NoError(t, rootCmd.Execute())
	}

	execCtx, cancel := context.WithCancel(context.Background())
	defer cancel() // always cancel the context to make sure the sub-process shuts down

	var callbackCount int
	testCallback := func() {
		callbackCount++
		t.Logf("Test callback called for the %dth time", callbackCount)
		currentBin, err := cfg.CurrentBin()
		require.NoError(t, err)
		switch callbackCount {
		case 1:
			// first startup
			// we should be starting with the genesis binary
			require.Contains(t, currentBin, "genesis")
			// add one manual upgrade
			go addManualUpgrade1()
		case 2:
			// first restart once we've add the first manual upgrade
			// ensure that the binary is still the genesis binary
			require.Contains(t, currentBin, "genesis")
			// add a second batch of manual upgrades
			go addManualUpgrade2()
		case 3:
			// next restart  after adding more manual upgrades
			// ensure that the binary is still the genesis binary
			require.Contains(t, currentBin, "genesis")
		case 4:
			// should have upgraded to manual10
			require.Contains(t, currentBin, "manual10")
		case 5:
			// should have upgraded to manual20
			require.Contains(t, currentBin, "manual20")
		case 6:
			// should have upgraded to gov1
			require.Contains(t, currentBin, "gov1")
		case 7:
			// should have upgraded to manual40
			require.Contains(t, currentBin, "manual40")
			// this is the end of our test so we shutdown after a bit here
			go func() {
				time.Sleep(pollInterval * 2)
				cancel()
			}()
		}
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"run", "--home", mockchainDir, "--cosmovisor-config", cfgFile})
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetErr(os.Stderr)
		execCtx = internal.WithTestCallback(execCtx, testCallback)
		require.NoError(t, rootCmd.ExecuteContext(execCtx))
		wg.Done()
	}()
	wg.Wait()

	require.Equal(t, 7, callbackCount)

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
